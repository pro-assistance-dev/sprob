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

	Children Fields `bun:"rel:has-many,join:id=parent_id" json:"children"`

	ParentID uuid.NullUUID `bun:"type:uuid" json:"parentId"`
	Parent   *Field        `bun:"-" json:"parent"`
}

type Fields []*Field

type FieldsWithCount struct {
	Fields Fields `json:"items"`
	Count  int    `json:"count"`
}
