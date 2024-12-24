package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FormSection struct {
	bun.BaseModel `bun:"form_sections,alias:form_sections"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	Fields        Fields        `bun:"rel:has-many" json:"fields"`

	Form   *Form         `bun:"rel:belongs-to" json:"form"`
	FormID uuid.NullUUID `bun:"type:uuid" json:"formId"`

	Order uint `bun:"item_order" json:"order"`
}

type FormSections []*FormSection

type FormSectionsWithCount struct {
	FormSections FormSections `json:"items"`
	Count        int          `json:"count"`
}
