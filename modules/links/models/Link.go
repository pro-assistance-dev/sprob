package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Link — управляемая ссылка: печатается обёртка /r/<code>, а цель живёт в БД.
// Смена домена/пути/маски не ломает напечатанные QR — правится цель, не код.
type Link struct {
	bun.BaseModel `bun:"links,alias:links"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	// Code — короткий неизменяемый идентификатор обёртки (без l/1/O/0).
	// ⚠️ `scanonly`: код генерит SQL-вставка (`InsertGenerated`). Без этого
	// авто-CRUD (`bun.NewInsert().Model`) писал ПУСТОЙ код — наклейка `//r/`.
	Code  string `bun:"link_code,unique,notnull,scanonly" json:"code"`
	Title string `json:"title"`

	// App — сервис, которому принадлежит ссылка: map | food | links.
	App string `bun:"app,notnull" json:"app"`
	// TargetType — тип цели: room/floor/building | diet/dish | survey | external.
	TargetType string `json:"targetType"`
	// TargetID — id сущности в её сервисе (для external не используется).
	TargetID string `json:"targetId"`
	// ExternalURL — абсолютный адрес для targetType=external.
	ExternalURL string `bun:"external_url" json:"externalUrl"`
	// Path — необязательный путь внутри сервиса (переопределяет путь по умолчанию).
	Path string `json:"path"`

	Enabled      bool       `bun:"enabled,notnull,default:true" json:"enabled"`
	RedirectKind int        `bun:"redirect_kind,default:302" json:"redirectKind"`
	ExpiresAt    *time.Time `bun:"expires_at" json:"expiresAt"`
	Note         string     `json:"note"`

	Hits int64 `bun:"hits,default:0" json:"hits"`

	CreatedAt time.Time `bun:"created_at,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at,default:current_timestamp" json:"updatedAt"`
}

func (item Link) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

// SearchColumns — поиск в списке (FTSP/options) по коду и подписи.
func (item *Link) SearchColumns() []string {
	return []string{"code", "title"}
}

type Links []*Link

type LinksWithCount struct {
	Links []*Link `json:"items"`
	Count int     `json:"count"`
}
