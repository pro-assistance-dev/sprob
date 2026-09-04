package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// AuditLog - журнал корректуры (Этап 3):
// кто/когда/какая сущность/какой объект/какое поле, старое -> новое.
type AuditLog struct {
	bun.BaseModel `bun:"audit_log,alias:audit_log"`
	Relationable

	CreatedAt time.Time     `json:"createdAt" bun:"created_at"`
	UserID    uuid.NullUUID `json:"userId,omitempty" bun:"user_id"`
	UserName  string        `json:"userName,omitempty" bun:"user_name"`
	RoleCode  string        `json:"roleCode,omitempty" bun:"role_code"`
	Entity    string        `json:"entity,omitempty" bun:"entity"`
	ObjectID  uuid.NullUUID `json:"objectId,omitempty" bun:"object_id"`
	// Инвентарный код объекта на момент правки
	Code      string `json:"code,omitempty" bun:"code"`
	Field     string `json:"field,omitempty" bun:"field"`
	OldValue  string `json:"oldValue,omitempty" bun:"old_value"`
	NewValue  string `json:"newValue,omitempty" bun:"new_value"`
	Operation string `json:"operation,omitempty" bun:"operation"` // create / update / delete
}

type AuditLogs []*AuditLog

type AuditLogsWithCount struct {
	AuditLogs AuditLogs `json:"items"`
	Count     int       `json:"count"`
}
