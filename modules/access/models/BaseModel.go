package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Relationable — базовая встраиваемая часть моделей модуля (id + Relation no-op),
// повторяет map/models.BaseModel (вынесено 04.09, С4.1).
type Relationable struct {
	ID uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
}

func (item Relationable) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
