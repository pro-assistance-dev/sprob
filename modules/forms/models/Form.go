package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Статусы формы (черновик/опубликована).
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

type Form struct {
	bun.BaseModel `bun:"forms,alias:forms" rus:"Форма"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name" rus:"Название"`

	// Status: draft | published. Опубликованная форма доступна для публикации наружу.
	Status string `json:"status"`

	// ImageURL — картинка-шапка формы (путь /api/static/... из file_infos).
	ImageURL string `json:"imageUrl"`

	FormSections FormSections `bun:"rel:has-many" json:"formSections"`

	Order uint `bun:"item_order" json:"order"`
}

type Forms []*Form

type FormsWithCount struct {
	Forms Forms `json:"items"`
	Count int   `json:"count"`
}

// SetIDForChildren проставляет id формы секциям и ниже по дереву.
// Детям без id генерирует uuid, чтобы FK можно было выставить до вставки
// (Bun не возвращает сгенерированные PK дочерних вложенных записей).
func (item *Form) SetIDForChildren() {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.FormSections.SetIDForChildren(item.ID)
}

func (items Forms) SetIDForChildren() {
	for i := range items {
		items[i].SetIDForChildren()
	}
}
