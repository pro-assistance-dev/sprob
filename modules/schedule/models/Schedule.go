package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Schedule — расписание одного места на один день (зал/кабинет на дату).
// Обобщённый аналог Schedule из pros: внутри секции (ScheduleSession)
// и слоты (ScheduleSlot).
//
// ⚠️ Таблица называется schedule_timetables (НЕ schedules): у некоторых проектов
// (portal, pros) уже есть свои таблицы/роуты `schedules` — коллизия. Роут тоже
// кастомный: /schedule-timetables.
type Schedule struct {
	bun.BaseModel `bun:"schedule_timetables,alias:schedule_timetables"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`

	ScheduleDayID uuid.NullUUID  `bun:"type:uuid,notnull" json:"scheduleDayId"`
	Day           *ScheduleDay   `bun:"rel:belongs-to,join:schedule_day_id=id" json:"day"`
	PlaceID       uuid.NullUUID  `bun:"type:uuid,notnull" json:"placeId"`
	Place         *SchedulePlace `bun:"rel:belongs-to" json:"place"`

	Sessions ScheduleSessions `bun:"rel:has-many,join:id=schedule_id" json:"sessions"`
	Slots    ScheduleSlots    `bun:"rel:has-many,join:id=schedule_id" json:"slots"`
}

type Schedules []*Schedule

type SchedulesWithCount struct {
	Schedules Schedules `json:"items"`
	Count     int       `json:"count"`
}

func (item Schedule) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		Relation("Day").
		Relation("Place").
		Relation("Sessions").
		Relation("Sessions.Slots").
		Relation("Slots")
}

