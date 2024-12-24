package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AnswerVariant struct {
	bun.BaseModel `bun:"answer_variants,alias:answer_variants"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	FieldID       uuid.NullUUID `bun:"type:uuid" json:"fieldId"`
	Field         *Field        `bun:"rel:belongs-to" json:"field"`
	Order         int           `bun:"item_order" json:"order"`
	Score         int           `json:"score"`

	ShowOther        bool               `json:"showOther"`
	AggregatedValues map[string]float64 `bun:"-" json:"aggregatedValues"`
	IsMatrixY        bool               `bun:"is_matrix_y" json:"isMatrixY"`
}

type AnswerVariants []*AnswerVariant

func (items AnswerVariants) GetRegisterPropertyOthersForDelete() []uuid.UUID {
	itemsForGet := make([]uuid.UUID, 0)
	//for i := range items {
	//	itemsForGet = append(itemsForGet, items[i].RegisterPropertyOthersForDelete...)
	//}
	return itemsForGet
}

func (item *AnswerVariant) writeXlsxAggregatedValues(key string) {
	_, ok := item.AggregatedValues[key]
	if ok {
		item.AggregatedValues[key]++
	} else {
		item.AggregatedValues[key] = 1
	}
}

//
// func (item *FieldFillVariant) GetAggregatedPercentage() {
// 	sum := float64(0)
// 	for k, v := range item.AggregatedValues {
// 		sum += v
// 		item.RegisterQueryPercentages = append(item.RegisterQueryPercentages, &ResearchQueryPercentage{k, v})
// 	}
// 	sort.Slice(item.RegisterQueryPercentages, func(i, j int) bool {
// 		return item.RegisterQueryPercentages[i].Value > item.RegisterQueryPercentages[j].Value
// 	})
// }
//
// func (items FieldFillVariants) Include(FieldFills FieldFills) string {
// 	exists := No
// 	for _, item := range items {
// 		if len(FieldFills) == 0 {
// 			break
// 		}
// 		for _, a := range FieldFills {
// 			for _, s := range a.SelectedAnswerVariants {
// 				if s.FieldFillVariantID == item.ID {
// 					exists = Yes
// 					break
// 				}
// 			}
// 			if exists == Yes {
// 				break
// 			}
// 		}
// 	}
// 	return exists
// }
