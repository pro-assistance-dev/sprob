package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Form struct {
	bun.BaseModel `bun:"forms,alias:forms" rus:"Форма"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name" rus:"Название"`

	FormSections FormSections `bun:"rel:has-many" json:"formSections"`

	Order uint `bun:"item_order" json:"order"`
}

type Forms []*Form

type FormsWithCount struct {
	Forms Forms `json:"items"`
	Count int   `json:"count"`
}
