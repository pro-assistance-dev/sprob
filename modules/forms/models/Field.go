package models

import (
	"github.com/google/uuid"
	basemodels "github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
)

type Field struct {
	bun.BaseModel `bun:"fields,alias:fields"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	ShortName     string        `json:"shortName"`
	Code          string        `json:"code"`

	Order    uint   `bun:"item_order" json:"order"`
	Comment  string `json:"comment"`
	Required bool   `json:"required"`
	// RequiredForCancel bool   `json:"requiredForCancel"`
	// Mask              string `json:"mask"`

	FormSection   *FormSection  `bun:"rel:belongs-to" json:"formSection"`
	FormSectionID uuid.NullUUID `bun:"type:uuid" json:"formSectionId"`

	ValueType   *basemodels.ValueType `bun:"rel:belongs-to" json:"valueType"`
	ValueTypeID uuid.NullUUID         `bun:"type:uuid" json:"valueTypeId"`

	// File   *FileInfo     `bun:"rel:belongs-to" json:"file"`
	// FileID uuid.NullUUID `bun:"type:uuid,nullzero,default:NULL" json:"fileId"`

	// MaskTokens          MaskTokens  `bun:"rel:has-many" json:"maskTokens"`
	// MaskTokensForDelete []uuid.UUID `bun:"-" json:"maskTokensForDelete"`
	AnswerVariants AnswerVariants `bun:"rel:has-many" json:"answerVariants"`

	// FieldRules — условия показа вопроса («если, то»): вопрос виден, только
	// когда выполнены ВСЕ правила (AND). Пусто — вопрос показывается всегда.
	FieldRules FieldRules `bun:"rel:has-many" json:"fieldRules"`

	Children Fields `bun:"rel:has-many,join:id=parent_id" json:"children"`

	ParentID uuid.NullUUID `bun:"type:uuid" json:"parentId"`
	Parent   *Field        `bun:"-" json:"parent"`
}

type Fields []*Field

type FieldsWithCount struct {
	Fields Fields `json:"items"`
	Count  int    `json:"count"`
}

// SetIDForChildren проставляет полю id секции и варианты детей/ответов/правил.
func (item *Field) SetIDForChildren(formSectionID uuid.NullUUID) {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.FormSectionID = formSectionID
	item.AnswerVariants.SetIDForChildren(item.ID)
	item.FieldRules.SetIDForChildren(item.ID)
	item.Children.SetIDForChildren(item.ID)
}

func (items Fields) SetIDForChildren(formSectionID uuid.NullUUID) {
	for i := range items {
		items[i].SetIDForChildren(formSectionID)
	}
}
