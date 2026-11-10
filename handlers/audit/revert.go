package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Режимы отката (10.11, замечание владельца: «отката должно быть два»).
//
//	single  — отменить ТОЛЬКО выбранное изменение: объект возвращается к
//	          состоянию «до» этой записи. Всё, что сделали ПОСЛЕ, остаётся.
//	cascade — отменить это изменение И ВСЁ, ЧТО БЫЛО ПОСЛЕ него у того же
//	          объекта.
//
// ⚠️ Почему это РАЗНЫЕ операции: при `single` отмена старого изменения может
// дать состояние, которого никогда не было (поменяли фамилию, потом отдел —
// отмена только фамилии оставит новый отдел со старой фамилией). `cascade`
// прокручивает историю объекта назад.
const (
	RevertSingle  = "single"
	RevertCascade = "cascade"
)

// RevertRequest — откат к состоянию из журнала.
type RevertRequest struct {
	ID   string `json:"id"`
	Mode string `json:"mode"`
}

// RevertResult — что произошло.
type RevertResult struct {
	Entity   string   `json:"entity"`
	Object   string   `json:"objectId"`
	Mode     string   `json:"mode"`
	Applied  []string `json:"applied"`
	Skipped  []string `json:"skipped"`
	Reverted int      `json:"reverted"`
}

// Revert — откат объекта по записи журнала.
//
//	POST <prefix>/revert  {"id":"<audit-row>","mode":"single|cascade"}
func (h *Handler) Revert(c *gin.Context) {
	var req RevertRequest
	if _, err := h.audit.helper.HTTP.GetForm(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректное тело запроса"})
		return
	}
	mode := req.Mode
	if mode == "" {
		mode = RevertSingle
	}
	if mode != RevertSingle && mode != RevertCascade {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode: ожидается single или cascade"})
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id записи журнала"})
		return
	}

	ctx := c.Request.Context()
	row, err := h.auditRow(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "запись журнала не найдена"})
		return
	}

	if row.OldValue == nil || strings.TrimSpace(*row.OldValue) == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "у записи нет состояния «было» — откат невозможен",
			"hint":  "журнал начал сохранять «было» 08.11.2026; записи старше этого не откатываются",
		})
		return
	}
	if row.ObjectID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "у записи нет объекта (object_id)"})
		return
	}

	result := RevertResult{Entity: row.Entity, Object: row.ObjectID, Mode: mode}

	if mode == RevertCascade {
		n, collected, skipped, cerr := h.revertCascade(ctx, row, id)
		if cerr != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": cerr.Error()})
			return
		}
		result.Applied = collected
		result.Skipped = skipped
		result.Reverted = n
	} else {
		// ⚠️ CREATE откатывается УДАЛЕНИЕМ созданной строки (снимка «до» у
		// создания нет — строка появилась этим действием).
		if row.Operation == "create" {
			if derr := h.undoCreate(ctx, row.Entity, row.ObjectID); derr != nil {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": derr.Error()})
				return
			}
			result.Applied = []string{"удалено"}
			result.Reverted = 1
		} else {
			applied, skipped, aerr := h.applySnapshot(ctx, row.Entity, row.ObjectID, *row.OldValue)
			if aerr != nil {
				h.audit.helper.HTTP.HandleError(c, aerr)
				return
			}
			if len(applied) == 0 {
				c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "нечего восстанавливать"})
				return
			}
			// ⚠️ После DELETE строка уже не существует: UPDATE задевает 0
			// строк, и «успех» был бы ложью. Вставляем строку из снимка.
			if row.Operation == "delete" {
				if inserted, ierr := h.restoreMissing(ctx, row.Entity, row.ObjectID, *row.OldValue); ierr == nil && inserted {
					applied = append(applied, "восстановлена строка")
				}
			}
			result.Applied = applied
			result.Skipped = skipped
			result.Reverted = 1
		}
	}

	// Сам откат — тоже изменение: пишем в журнал, иначе «кто вернул» потеряется.
	mark, _ := json.Marshal(map[string]interface{}{
		"revertedFrom": req.ID,
		"mode":         mode,
		"count":        result.Reverted,
	})
	markStr := string(mark)
	_, _ = h.audit.helper.DB.IDB(ctx).NewRaw(`
		INSERT INTO audit_log (entity, object_id, operation, old_value, new_value, user_name)
		VALUES (?, ?, 'update', ?, ?, ?)`,
		row.Entity, row.ObjectID, row.OldValue, markStr, actorName(c)).Exec(ctx)

	c.JSON(http.StatusOK, result)
}

// revertCascade — отменить выбранную запись и ВСЕ более новые по этому объекту.
//
// ⚠️ Идём ОТ НОВЫХ К СТАРЫМ и применяем каждый снимок `old_value`; последним
// ложится снимок «до» самой выбранной записи — это и есть состояние «как было
// до выбранного момента».
//
// ⚠️ Записи-«откаты» (в `new_value` есть `revertedFrom`) в цепочку НЕ берём:
// их снимок — состояние на момент отката, и применение «поверх» рушит результат
// (после `single` ШАГ2 → ШАГ1 `cascade` возвращал ШАГ2, проверено 10.11).
//
// ⚠️ Границу берём по `created_at`, а не по id: id случайный (uuid_v4).
func (h *Handler) revertCascade(
	ctx context.Context,
	row auditRowData,
	fromID uuid.UUID,
) (reverted int, applied []string, skipped []string, err error) {
	if _, ok := h.audit.options.RevertTables[row.Entity]; !ok {
		return 0, nil, nil, fmt.Errorf("для сущности «%s» откат не поддержан", row.Entity)
	}

	type chainItem struct {
		ID       uuid.UUID `bun:"id"`
		OldValue *string   `bun:"old_value"`
	}
	items := make([]chainItem, 0)
	const q = `
		SELECT id, old_value FROM audit_log
		WHERE object_id = ?
		  AND created_at >= (SELECT created_at FROM audit_log WHERE id = ?)
		  AND old_value IS NOT NULL AND old_value <> ''
		  AND coalesce(new_value, '') NOT LIKE '%"revertedFrom"%'
		ORDER BY created_at DESC`
	if err = h.audit.helper.DB.IDB(ctx).NewRaw(q, row.ObjectID, fromID).Scan(ctx, &items); err != nil {
		return 0, nil, nil, err
	}
	if len(items) == 0 {
		return 0, nil, nil, fmt.Errorf("нечего откатывать: у записей нет состояния «было»")
	}

	seen := map[string]bool{}
	skipSeen := map[string]bool{}
	for _, it := range items {
		a, s, aerr := h.applySnapshot(ctx, row.Entity, row.ObjectID, *it.OldValue)
		if aerr != nil {
			return reverted, applied, skipped, aerr
		}
		for _, col := range a {
			if !seen[col] {
				seen[col] = true
				applied = append(applied, col)
			}
		}
		for _, col := range s {
			if !skipSeen[col] {
				skipSeen[col] = true
				skipped = append(skipped, col)
			}
		}
		reverted++
	}
	return reverted, applied, skipped, nil
}

// applySnapshot — записать снимок в строку таблицы.
//
// ⚠️ Колонки фильтруем по РЕАЛЬНОЙ схеме: снимок мог быть снят на другой
// версии кода, и «лишняя» колонка уронила бы весь откат.
func (h *Handler) applySnapshot(
	ctx context.Context,
	entity, objectID, oldValue string,
) (applied []string, skipped []string, err error) {
	table, ok := h.audit.options.RevertTables[entity]
	if !ok {
		return nil, nil, fmt.Errorf("для сущности «%s» откат не поддержан", entity)
	}

	var snap map[string]interface{}
	if err = json.Unmarshal([]byte(oldValue), &snap); err != nil {
		return nil, nil, fmt.Errorf("состояние «было» повреждено: %w", err)
	}

	existing, err := h.audit.tableColumns(ctx, table)
	if err != nil {
		return nil, nil, err
	}

	q := h.audit.helper.DB.IDB(ctx).NewUpdate().Table(table)
	set := false
	applied = make([]string, 0, len(snap))
	skipped = make([]string, 0)
	for col, val := range snap {
		if col == "id" || !existing[col] {
			skipped = append(skipped, col)
			continue
		}
		q = q.Set(`"`+col+`" = ?`, val)
		applied = append(applied, col)
		set = true
	}
	if !set {
		return nil, skipped, nil
	}
	if _, err = q.Where("id = ?", objectID).Exec(ctx); err != nil {
		return nil, nil, err
	}
	return applied, skipped, nil
}
