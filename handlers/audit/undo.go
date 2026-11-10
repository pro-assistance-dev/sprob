package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Откат CREATE и восстановление удалённых строк.
//
// ⚠️ ЗАЧЕМ (10.11). В портале журнал умел больше, чем в hr: отменять СОЗДАНИЕ
// (удалить созданную строку) и восстанавливать УДАЛЁННУЮ (INSERT из снимка,
// если строки уже нет). Перенос журнала в sprob не должен ничего терять —
// поэтому обе операции здесь. Для UPDATE/удаления-если-есть хватает снимка.

// undoCreate — отменить создание: удалить строку, созданную этим действием.
func (h *Handler) undoCreate(ctx context.Context, entity, objectID string) error {
	if objectID == "" {
		return fmt.Errorf("нет id созданной записи")
	}
	table, ok := h.audit.options.RevertTables[entity]
	if !ok {
		// Сущность есть в снимке, но откат не заявлен — ищем таблицу по карте снимков.
		table, ok = snapshotTables[entity]
	}
	if !ok {
		return fmt.Errorf("таблица для сущности %q не найдена", entity)
	}
	_, err := h.audit.helper.DB.IDB(ctx).NewRaw(
		`DELETE FROM `+table+` WHERE id = ?`, objectID,
	).Exec(ctx)
	return err
}

// restoreMissing — вставить строку из снимка, ЕСЛИ её больше нет.
//
// ⚠️ Зачем: удаление откатывается не только «вернуть значения» (UPDATE), но и
// «вернуть саму строку» (INSERT): после delete UPDATE задевает 0 строк.
// Возвращает true, если строка была вставлена.
func (h *Handler) restoreMissing(
	ctx context.Context,
	entity, objectID, oldValue string,
) (bool, error) {
	table, ok := h.audit.options.RevertTables[entity]
	if !ok {
		table, ok = snapshotTables[entity]
	}
	if !ok {
		return false, fmt.Errorf("таблица для сущности %q не найдена", entity)
	}

	var snap map[string]interface{}
	if err := json.Unmarshal([]byte(oldValue), &snap); err != nil {
		return false, fmt.Errorf("состояние «было» повреждено: %w", err)
	}
	if _, hasID := snap["id"]; !hasID {
		snap["id"] = objectID
	}

	existing, err := h.audit.tableColumns(ctx, table)
	if err != nil {
		return false, err
	}

	names := make([]string, 0, len(snap))
	phs := make([]string, 0, len(snap))
	args := make([]interface{}, 0, len(snap))
	for col, val := range snap {
		if !existing[col] {
			continue
		}
		names = append(names, `"`+col+`"`)
		phs = append(phs, "?")
		args = append(args, val)
	}
	if len(names) == 0 {
		return false, fmt.Errorf("в снимке нет вставляемых колонок")
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
		table, strings.Join(names, ", "), strings.Join(phs, ", "))
	_, err = h.audit.helper.DB.IDB(ctx).NewRaw(q, args...).Exec(ctx)
	return err == nil, err
}
