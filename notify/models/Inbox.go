package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Inbox — входящее уведомление «в приложении» (in-app): адресовано конкретному
// пользователю (`UserID`, из Keycloak sub / ClaimUserID) либо роли/группе.
// Каналы email/telegram/webhook — «выстрелил и забыл»; in-app хранится до
// прочтения (`ReadAt`), чтобы клиент мог показать бейдж/колокольчик.
type Inbox struct {
	bun.BaseModel `bun:"notify_inbox,alias:notify_inbox"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	// UserID — получатель (id пользователя). Пусто — адресовано роли/группе.
	UserID string `bun:"user_id" json:"userId"`
	// Group — группа/роль получателя (напр. «admin»), если UserID пуст.
	Group string `json:"group"`

	Title string `json:"title"`
	Body  string `bun:"type:text" json:"body"`
	// URL — ссылка (маршрут в SPA), куда ведёт уведомление.
	URL string `json:"url"`

	RuleID uuid.NullUUID `bun:"type:uuid" json:"ruleId"`
	Event  string        `json:"event"`
	Entity string        `json:"entity"`
	ItemID string        `json:"itemId"`

	// ReadAt — когда прочитано; nil — новое (для бейджа).
	ReadAt    *time.Time `bun:"read_at" json:"readAt,omitempty"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:now()" json:"createdAt"`
}

type InboxItems []*Inbox

type InboxWithCount struct {
	Items   InboxItems `json:"items"`
	Unread  int        `json:"unread"`
	Count   int        `json:"count"`
}
