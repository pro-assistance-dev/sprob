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

	// Description — ВВОДНЫЙ ТЕКСТ над полями: даты, правила, куда писать.
	//
	// ⚠️ Это НЕ поле формы (`Fields`). Поле ввода имеет значение ответа и
	// попадает в `field_fills`; описание — статический текст, ответа не
	// имеет. Если сделать его полем, оно попало бы в выгрузки и аналитику
	// как «вопрос без ответа».
	//
	// ⚠️ Показывается на публичной странице между заголовком и первым
	// вопросом, и правится в конструкторе — один раз, не в каждом проекте.
	Description string `bun:"description,notnull,default:''" json:"description"`

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
