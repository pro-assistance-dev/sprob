package access

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pro-assistance-dev/sprob/modules/access/models"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun"
)

// auditRow - одна запись журнала корректуры.
type auditRow struct {
	Entity    string
	ObjectID  string
	Code      string
	Field     string
	OldValue  string
	NewValue  string
	Operation string
}

// auditCreate фиксирует создание: по строке на поле (old = "").
// Пишутся только реальные скалярные поля сущности — без дампов объектов.
func (m *Middleware) auditCreate(ctx context.Context, entity *Entity, id string, payload map[string]interface{}, uc UserCtx) {
	rows := make([]auditRow, 0, len(payload))
	for field, val := range payload {
		if !isAuditableField(entity, field, val) {
			continue
		}
		rows = append(rows, auditRow{
			Entity: entity.Key, ObjectID: id,
			Code: objectCode(payload), Field: field,
			NewValue: stringify(val), Operation: "create",
		})
	}
	m.writeAudit(ctx, rows, uc)
}

// auditUpdate фиксирует изменение: только поля, значения которых реально изменились.
// oldRow - состояние строки ДО правки (читается middleware'ом до handler'а).
// Пишутся только реальные скалярные поля сущности: объекты/массивы не дампим
// (для связей есть *_id, который логируется отдельно).
func (m *Middleware) auditUpdate(ctx context.Context, entity *Entity, id string, oldRow map[string]interface{}, payload map[string]interface{}, uc UserCtx) {
	rows := make([]auditRow, 0, len(payload))
	for field, newVal := range payload {
		if !isAuditableField(entity, field, newVal) {
			continue
		}
		oldVal := ""
		if oldRow != nil {
			// bun сканирует в map с ключами-колонками: json-имя -> snake_case;
			// для связей (department, worker, ...) колонка имеет суффикс _id
			col := strcase.ToSnake(field)
			if v, ok := oldRow[col]; ok {
				oldVal = stringify(v)
			} else if v, ok := oldRow[col+"_id"]; ok {
				oldVal = stringify(v)
			}
		}
		newStr := stringify(newVal)
		if oldVal == newStr {
			continue
		}
		rows = append(rows, auditRow{
			Entity: entity.Key, ObjectID: id,
			Code: objectCode(payload), Field: field,
			OldValue: oldVal, NewValue: newStr, Operation: "update",
		})
	}
	m.writeAudit(ctx, rows, uc)
}

// auditDelete фиксирует удаление объекта (без пофайлового разбора).
func (m *Middleware) auditDelete(ctx context.Context, entity *Entity, id string, payload map[string]interface{}, uc UserCtx) {
	rows := []auditRow{{
		Entity: entity.Key, ObjectID: id,
		Code: objectCode(payload), Operation: "delete",
	}}
	m.writeAudit(ctx, rows, uc)
}

// fetchOldRow читает текущее состояние строки до правки (для diff).
func (m *Middleware) fetchOldRow(ctx context.Context, entity *Entity, id string) map[string]interface{} {
	if id == "" || entity.Table == "" || m.helper == nil {
		return nil
	}
	row := map[string]interface{}{}
	err := m.helper.DB.IDB(ctx).NewSelect().
		TableExpr("?0", bun.Ident(entity.Table)).
		Where("id = ?", id).
		Scan(ctx, &row)
	if err != nil {
		return nil
	}
	return row
}

// auditWriteErrors - счётчик ошибок записи журнала корректуры (В6): потери
// записей должны быть заметны (метрика/лог), а не только log.Printf.
var auditWriteErrors atomic.Int64

// AuditWriteErrors возвращает число неудачных записей журнала корректуры.
func AuditWriteErrors() int64 { return auditWriteErrors.Load() }

func (m *Middleware) writeAudit(ctx context.Context, rows []auditRow, uc UserCtx) {
	if len(rows) == 0 {
		return
	}
	// Контекст запроса после завершения handler'а может быть отменён,
	// поэтому пишем аудит в контексте, независимом от отмены.
	wctx := context.WithoutCancel(ctx)
	userID := uuid.NullUUID{}
	if uc.UserID != "" {
		if parsed, err := uuid.Parse(uc.UserID); err == nil {
			userID = uuid.NullUUID{UUID: parsed, Valid: true}
		}
	}
	role := ""
	if len(uc.Roles) > 0 {
		role = uc.Roles[0]
	}
	userName := uc.UserName
	if userName == "" {
		// фоновые записи (импорты, datasync) и неавторизованные сессии
		userName = "система"
	}
	logs := make(models.AuditLogs, 0, len(rows))
	for _, r := range rows {
		objectID := uuid.NullUUID{}
		if r.ObjectID != "" {
			if parsed, err := uuid.Parse(r.ObjectID); err == nil {
				objectID = uuid.NullUUID{UUID: parsed, Valid: true}
			}
		}
		logs = append(logs, &models.AuditLog{
			CreatedAt: time.Now(),
			UserID:    userID,
			UserName:  userName,
			RoleCode:  role,
			Entity:    r.Entity,
			ObjectID:  objectID,
			Code:      r.Code,
			Field:     r.Field,
			OldValue:  truncate(r.OldValue),
			NewValue:  truncate(r.NewValue),
			Operation: r.Operation,
		})
	}
	if _, err := m.helper.DB.IDB(wctx).NewInsert().Model(&logs).Exec(wctx); err != nil {
		n := auditWriteErrors.Add(1)
		log.Printf("[audit] writeAudit ERROR #%d: %v (rows=%d)", n, err, len(logs))
	}
}

// objectCode извлекает инвентарный код объекта из payload.
func objectCode(payload map[string]interface{}) string {
	for _, key := range []string{"fullCode", "objectCode", "code", "hostCode", "number"} {
		if v, ok := payload[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// isAuditableField - поле участвует в журнале, только если:
//  1. это реальное поле сущности (реестр access/registry.go), а не служебная колонка (search и т.п.);
//  2. значение скалярное (объекты/массивы не дампим — для связей логируется *_id).
func isAuditableField(entity *Entity, field string, val interface{}) bool {
	if field == "id" {
		return false
	}
	if _, ok := entity.Fields[field]; !ok {
		return false
	}
	switch val.(type) {
	case map[string]interface{}, []interface{}:
		return false
	}
	return true
}

// stringify приводит значение к строке для журнала: UUID-колонки из БД приходят
// как []byte (иначе получаем base64), числа/булевы — человекочитаемо.
func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case bool:
		return fmt.Sprintf("%v", t)
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case time.Time:
		return t.Format("2006-01-02 15:04:05")
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func truncate(s string) string {
	const maxLen = 2000
	if len(s) > maxLen {
		return strings.TrimSpace(s[:maxLen]) + "..."
	}
	return s
}

// findCreatedID ищет id только что созданной строки по значениям payload.
// Нужен, когда POST пришёл без id в payload: sprob Create делает
// Insert без RETURNING, поэтому в ответе id=null и взять его неоткуда.
// Используются только скалярные поля сущности из payload (не пустые).
func (m *Middleware) findCreatedID(ctx context.Context, entity *Entity, payload map[string]interface{}) string {
	if entity.Table == "" || len(payload) == 0 {
		return ""
	}
	q := m.helper.DB.IDB(ctx).NewSelect().
		TableExpr("?0", bun.Ident(entity.Table)).
		Column("id")
	matched := false
	for field, val := range payload {
		if !isAuditableField(entity, field, val) {
			continue
		}
		if isEmptyValue(val) {
			continue
		}
		q = q.Where("?0 = ?", bun.Ident(strcase.ToSnake(field)), val)
		matched = true
	}
	if !matched {
		return ""
	}
	q = q.Limit(1)
	var id string
	if err := q.Scan(ctx, &id); err != nil {
		return ""
	}
	return id
}

// isEmptyValue - пустое значение поля, не участвующее в поиске созданной строки.
func isEmptyValue(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	case bool:
		return !t
	case float64:
		return t == 0
	}
	return false
}
