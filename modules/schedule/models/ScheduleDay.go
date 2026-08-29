package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ScheduleDay — день календаря-расписания (обобщённый аналог EventDay из pros).
// Привязывается к любому владельцу через OwnerID/OwnerType
// (мероприятие, врач, конференция — что угодно).
type ScheduleDay struct {
	bun.BaseModel `bun:"schedule_days,alias:schedule_days"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Date          time.Time     `bun:"item_date" json:"date"`
	Description   string        `json:"description"`

	// Владелец расписания: например eventId (конференция) или doctorId (портал больницы).
	OwnerID   uuid.NullUUID `bun:"type:uuid,notnull" json:"ownerId"`
	OwnerType string        `bun:"type:varchar,notnull" json:"ownerType"`

	Schedules Schedules `bun:"rel:has-many,join:id=schedule_day_id" json:"schedules"`
}

type ScheduleDays []*ScheduleDay

type ScheduleDaysWithCount struct {
	ScheduleDays ScheduleDays `json:"items"`
	Count        int          `json:"count"`
}

// Relation подгружает всё дерево календаря одним запросом:
// день → расписания (залы/кабинеты) → места, секции и слоты.
func (item ScheduleDay) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		Relation("Schedules").
		Relation("Schedules.Place").
		Relation("Schedules.Sessions").
		Relation("Schedules.Sessions.Slots").
		Relation("Schedules.Slots")
}

