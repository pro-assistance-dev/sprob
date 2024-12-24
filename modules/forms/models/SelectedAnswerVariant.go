package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SelectedAnswerVariant struct {
	bun.BaseModel `bun:"selected_answer_variants,alias:selected_answer_variants"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	// Answer     *FieldFill    `bun:"rel:belongs-to" json:"fieldFill"`
	FieldFillId uuid.NullUUID `bun:"type:uuid" json:"fieldFillId"`

	// AnswerVariant   *AnswerVariant `bun:"rel:belongs-to" json:"answerVariant"`

	AnswerVariantID  uuid.NullUUID `bun:"type:uuid" json:"answerVariantId"`
	AnswerVariantIDY uuid.NullUUID `bun:"answer_variant_id_y,type:uuid" json:"answerVariantIdY"`
}

type SelectedAnswerVariants []*SelectedAnswerVariant

func (item *SelectedAnswerVariant) SetIDForChildren() {
	//if len(item.RegisterPropertyOthers) == 0 {
	//	return
	//}
	//for i := range item.RegisterPropertyOthers {
	//	item.RegisterPropertyOthers[i].AnswerVariantID = item.ID
	//}
}

func (items SelectedAnswerVariants) SetIDForChildren() {
	if len(items) == 0 {
		return
	}
	for i := range items {
		items[i].SetIDForChildren()
	}
}
