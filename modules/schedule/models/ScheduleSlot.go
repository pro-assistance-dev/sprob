package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ScheduleSlot — слот расписания: событие с временным интервалом
// (обобщённый аналог Perfom из pros: доклад, приём врача, лекция...).
//
// Payload — JSONB-«карман» для доменных данных проекта:
// спикеры, спонсор, формат, специализация врача и т.п. —
// чтобы модуль оставался универсальным без миграций под каждую предметку.
type ScheduleSlot struct {
	bun.BaseModel `bun:"schedule_slots,alias:schedule_slots"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	StartTime     string        `json:"startTime"`
	EndTime       string        `json:"endTime"`

	ScheduleID uuid.NullUUID `bun:"type:uuid,notnull" json:"scheduleId"`
	SessionID  uuid.NullUUID `bun:"type:uuid" json:"sessionId"`

	Payload map[string]interface{} `bun:"type:jsonb,jsonb" json:"payload"`
}

type ScheduleSlots []*ScheduleSlot

type ScheduleSlotsWithCount struct {
	ScheduleSlots ScheduleSlots `json:"items"`
	Count         int           `json:"count"`
}

func (item ScheduleSlot) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

