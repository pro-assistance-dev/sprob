package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"roles,alias:roles"`
	Relationable

	Code         string     `json:"code"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Periodicity  string     `json:"periodicity"`
	Consequences string     `json:"consequences"`
	Responsible  string     `json:"responsible"`
	Status       string     `json:"status"`
	DisbandedAt  *time.Time `bun:"disbanded_at" json:"disbandedAt"`
}

type Roles []*Role

type RolesWithCount struct {
	Roles Roles `json:"items"`
	Count int   `json:"count"`
}
