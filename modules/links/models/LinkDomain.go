package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// LinkDomain — хост, на котором живут обёртки. Активный домен один
// (IsDefault=true); историю не удаляем, чтобы понимать, какие наклейки ещё
// указывают на старый хост.
type LinkDomain struct {
	bun.BaseModel `bun:"link_domains,alias:link_domains"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Host      string `bun:"host,unique,notnull" json:"host"`
	Scheme    string `bun:"scheme,notnull,default:'https'" json:"scheme"`
	IsDefault bool   `bun:"is_default,default:false" json:"isDefault"`
	Note      string `json:"note"`
}

func (item LinkDomain) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (item *LinkDomain) SearchColumns() []string {
	return []string{"host", "note"}
}

type LinkDomains []*LinkDomain

type LinkDomainsWithCount struct {
	LinkDomains []*LinkDomain `json:"items"`
	Count       int           `json:"count"`
}

// LinkHit — переход по обёртке. Пишется асинхронно (очередь), чистится по
// LINK_HIT_RETENTION_DAYS (по умолчанию 180 дней, решение 18.10).
type LinkHit struct {
	bun.BaseModel `bun:"link_hits,alias:link_hits"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	LinkID    uuid.NullUUID `bun:"link_id,type:uuid" json:"linkId"`
	Code      string        `bun:"code" json:"code"`
	IP        string        `bun:"ip" json:"ip"`
	UserAgent string        `bun:"user_agent" json:"userAgent"`
	Referer   string        `bun:"referer" json:"referer"`
	CreatedAt time.Time     `bun:"created_at,default:current_timestamp" json:"createdAt"`
}

func (item LinkHit) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

type LinkHits []*LinkHit
