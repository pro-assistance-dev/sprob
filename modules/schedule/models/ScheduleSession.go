package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ScheduleSession — секция/группа расписания: временной интервал
// с вложенными слотами (обобщённый аналог Session из pros).
type ScheduleSession struct {
	bun.BaseModel `bun:"schedule_sessions,alias:schedule_sessions"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	StartTime     string        `json:"startTime"`
	EndTime       string        `json:"endTime"`

	ScheduleID uuid.NullUUID `bun:"type:uuid,notnull" json:"scheduleId"`

	Slots ScheduleSlots `bun:"rel:has-many,join:id=session_id" json:"slots"`
}

type ScheduleSessions []*ScheduleSession

type ScheduleSessionsWithCount struct {
	ScheduleSessions ScheduleSessions `json:"items"`
	Count            int              `json:"count"`
}

func (item ScheduleSession) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q.Relation("Slots")
}

