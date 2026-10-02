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

// SetIDForChildren проставляет секции id формы и её полям — id секции.
func (item *FormSection) SetIDForChildren(formID uuid.NullUUID) {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.FormID = formID
	item.Fields.SetIDForChildren(item.ID)
}

func (items FormSections) SetIDForChildren(formID uuid.NullUUID) {
	for i := range items {
		items[i].SetIDForChildren(formID)
	}
}
