package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FieldFill struct {
	bun.BaseModel `bun:"field_fills,alias:field_fills"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `

	ValueString string    `json:"valueString"`
	ValueNumber float32   `json:"valueNumber"`
	ValueDate   time.Time `json:"valueDate"`
	ValueOther  string    `json:"valueOther"`
	// FieldFillVariant   *FieldFillVariant `bun:"rel:belongs-to" json:"FieldFillVariant"`
	// FieldFillVariantID uuid.NullUUID  `bun:"type:uuid" json:"FieldFillVariantId"`

	FormFill   *FormFill     `bun:"rel:belongs-to" json:"formFill"`
	FormFillID uuid.NullUUID `bun:"type:uuid" json:"formFillId"`
	Filled     bool          `json:"filled"`

	Field   *Field        `bun:"rel:belongs-to" json:"field"`
	FieldID uuid.NullUUID `bun:"type:uuid" json:"fieldId"`

	// FieldVariant   *FieldVariant `bun:"rel:belongs-to" json:"FieldVariant"`
	AnswerVariantId uuid.NullUUID `bun:"type:uuid" json:"answerVariantId"`

	SelectedAnswerVariants SelectedAnswerVariants `bun:"rel:has-many" json:"selectedAnswerVariants"`
	// FieldFillFilesForDelete []uuid.UUID `bun:"-" json:"FieldFillFilesForDelete"`
}

type FieldFills []*FieldFill

type FieldFillsWithCount struct {
	FieldFills FieldFills `json:"items"`
	Count      int        `json:"count"`
}

// func (items FieldFills) GetFieldFillFiles() FieldFillFiles {
// 	itemsForGet := make(FieldFillFiles, 0)
// 	if len(items) == 0 {
// 		return itemsForGet
// 	}
// 	for i := range items {
// 		itemsForGet = append(itemsForGet, items[i].FieldFillFiles...)
// 	}
//
// 	return itemsForGet
// }

const (
	Yes    string = "Да"
	No     string = "Нет"
	NoData string = "Нет данных"
)

//
// func (items FieldFills) GetFieldFillFilesForDelete() []uuid.UUID {
// 	itemsForGet := make([]uuid.UUID, 0)
// 	if len(items) == 0 {
// 		return itemsForGet
// 	}
// 	for i := range items {
// 		itemsForGet = append(itemsForGet, items[i].FieldFillFilesForDelete...)
// 	}
//
// 	return itemsForGet
// }

func (item *FieldFill) GetData(q *Field) interface{} {
	if q.ValueType.IsString() || q.ValueType.IsText() {
		return item.ValueString
	}
	if q.ValueType.IsNumber() {
		return item.ValueNumber
	}
	if q.ValueType.IsDate() {
		return item.ValueDate
	}
	if q.ValueType.IsRadio() {
		res := No
		// for _, radio := range q.FieldFillVariants {
		// 	if radio.ID == item.FieldFillVariantID {
		// 		res = radio.Name
		// 		break
		// 	}
		// }
		return res
	}
	if q.ValueType.IsSet() {
		res := ""
		// for _, v := range item.SelectedAnswerVariants {
		// 	res += v.FieldFillVariant.Name + "; "
		// }
		// for _, radio := range q.FieldFillVariants {
		// if radio.ID == item.FieldFillVariantID {
		// 	res = radio.Name
		// 	break
		// }
		// }
		return res
	}
	return ""
}

func (item *FieldFill) GetAggregateExistingData() bool {
	if item.Field.ValueType.IsString() || item.Field.ValueType.IsText() {
		return len(item.ValueString) > 0
	}
	if item.Field.ValueType.IsNumber() {
		return item.ValueNumber > 0
	}
	if item.Field.ValueType.IsDate() {
		return !item.ValueDate.IsZero()
	}
	// if item.Field.ValueType.IsRadio() {
	// 	if item.FieldFillVariantID.Valid {
	// 		return true
	// 	}
	// }
	return false
}

func (item *FieldFill) FieldFillVariantSelected(variantID uuid.NullUUID) string {
	res := No
	// for _, selectedVariant := range item.SelectedAnswerVariants {
	// 	if selectedVariant.FieldFillVariantID == variantID {
	// 		res = Yes
	// 		break
	// 	}
	// }
	return res
}
