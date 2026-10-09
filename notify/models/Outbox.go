package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Meta — произвольные параметры правила (URL/Secret вебхука, chat_id и т.п.).
type Meta map[string]string

// Outbox — очередь исходящих уведомлений: переживает рестарт сервера и сбои
// транспорта. Воркер `notify.Notifier` отправляет строки с sent_at IS NULL и
// наступившим next_attempt_at, увеличивая attempts и откладывая повтор (backoff).
type Outbox struct {
	bun.BaseModel `bun:"notify_outbox,alias:notify_outbox"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Channel   string `json:"channel"`
	ToAddress string `bun:"to_address,notnull" json:"toAddress"`
	Subject   string `json:"subject"`
	Body      string `bun:"type:text,notnull" json:"body"`
	// Meta — параметры канала на момент постановки (URL/Secret вебхука и т.п.),
	// чтобы отправка не зависела от последующего изменения правила.
	Meta Meta `bun:"meta,type:jsonb" json:"meta"`

	RuleID uuid.NullUUID `bun:"type:uuid" json:"ruleId"`
	Event  string        `json:"event"`
	Entity string        `json:"entity"`
	ItemID string        `json:"itemId"`

	Attempts      int       `bun:"attempts,notnull,default:0" json:"attempts"`
	NextAttemptAt time.Time `bun:"next_attempt_at,notnull" json:"nextAttemptAt"`
	SentAt        *time.Time `bun:"sent_at" json:"sentAt,omitempty"`
	LastError     string    `bun:"last_error,type:text" json:"lastError,omitempty"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:now()" json:"createdAt"`
}

type OutboxItems []*Outbox
