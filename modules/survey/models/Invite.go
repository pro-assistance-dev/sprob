package models

import (
	"time"

	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

// Invite — адресное приглашение на неанонимную публикацию: персональная ссылка
// с токеном. Позволяет ограничить круг респондентов (сотрудники/пациенты) и
// учесть, кто уже ответил. Для анонимных публикаций приглашения не нужны.
type Invite struct {
	bun.BaseModel `bun:"invites,alias:invites"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Publication   *Publication  `bun:"rel:belongs-to" json:"publication"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`

	// Token — секрет персональной ссылки (`/f/<slug>?invite=<token>`). Уникален.
	Token string `json:"token"`
	// Email/Phone/Fio — подсказки адресата (PII, показываются только в админке).
	Email string `json:"email"`
	Phone string `json:"phone"`
	Fio   string `json:"fio"`

	// UsedAt — когда по приглашению ответили (одноразовое).
	UsedAt    *time.Time `bun:"used_at" json:"usedAt"`
	ExpiresAt *time.Time `bun:"expires_at" json:"expiresAt"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type Invites []*Invite

type InvitesWithCount struct {
	Invites Invites `json:"items"`
	Count   int     `json:"count"`
}

func (item Invite) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

// IsUsable — можно ли ответить по приглашению (не использовано, не истекло).
func (item *Invite) IsUsable() bool {
	if item.UsedAt != nil {
		return false
	}
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return false
	}
	return true
}
