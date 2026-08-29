package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// SchedulePlace — место проведения (зал, аудитория, кабинет врача и т.п.).
// Обобщённый аналог Place из pros.
type SchedulePlace struct {
	bun.BaseModel `bun:"schedule_places,alias:schedule_places"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Color         string        `json:"color"`
}

type SchedulePlaces []*SchedulePlace

type SchedulePlacesWithCount struct {
	SchedulePlaces SchedulePlaces `json:"items"`
	Count          int            `json:"count"`
}

func (item SchedulePlace) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

