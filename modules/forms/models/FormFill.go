package models

import (
	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

type FormFill struct {
	bun.BaseModel `bun:"form_fills,alias:form_fills"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	// Date          *time.Time    `bun:"item_date" json:"date"`

	Form   *Form         `bun:"rel:belongs-to" json:"form"`
	FormID uuid.NullUUID `bun:"type:uuid" json:"formId"`

	FormFills FormFills `bun:"rel:has-many" json:"formFills"`

	// CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"createdAt"`

	FieldFills FieldFills `bun:"rel:has-many" json:"fieldFills"`

	FillingPercentage uint `json:"fillingPercentage"`
	// Order             uint `bun:"item_order" json:"order"`
}

type FormFills []*FormFill

type FormFillsWithCount struct {
	FormFills FormFills `json:"items"`
	Count     int       `json:"count"`
}

func (items FormFills) GetLastResult() *FormFill {
	var lastResult *FormFill
	// for i := range items {
	// if lastResult == nil || lastResult.Date.Before(*items[i].Date) {
	// 	lastResult = items[i]
	// }
	// }
	return lastResult
}

func (item *FormFill) SetIDForChildren() {
	//if len(item.RegisterPropertyToPatient) > 0 {
	//	for i := range item.RegisterPropertyToPatient {
	//		item.RegisterPropertyToPatient[i].FormFillID = item.ID
	//	}
	//}
	if len(item.FieldFills) > 0 {
		for i := range item.FieldFills {
			item.FieldFills[i].FormFillID = item.ID
		}
	}
	//if len(item.RegisterPropertyOthersToPatient) > 0 {
	//	for i := range item.RegisterPropertyOthersToPatient {
	//		item.RegisterPropertyOthersToPatient[i].FormFillID = item.ID
	//	}
	//}
}

func (items FormFills) SetIDForChildren() {
	for i := range items {
		items[i].SetIDForChildren()
	}
}

func (items FormFills) SetDeleteIDForChildren() {
	//for i := range items {
	//	items[i].RegisterPropertyToPatientForDelete = append(items[i].RegisterPropertyToPatientForDelete, item.FieldFill[i].ID)
	//}
	//for i := range item.FieldFill {
	//	item.RegisterPropertySetToPatientForDelete = append(item.RegisterPropertySetToPatientForDelete, item.FieldFill[i].ID)
	//}
}

func (items FormFills) GetRegisterPropertiesToPatients() FieldFills {
	itemsForGet := make(FieldFills, 0)
	if len(items) == 0 {
		return itemsForGet
	}
	//for i := range items {
	//	itemsForGet = append(itemsForGet, items[i].RegisterPropertyToPatient...)
	//}

	return itemsForGet
}

func (items FormFills) GetFieldFills() FieldFills {
	itemsForGet := make(FieldFills, 0)
	if len(items) == 0 {
		return itemsForGet
	}
	return itemsForGet
}

func (items FormFills) GetRegisterPropertiesToPatientsForDelete() []uuid.UUID {
	itemsForGet := make([]uuid.UUID, 0)
	if len(items) == 0 {
		return itemsForGet
	}
	//for i := range items {
	//	itemsForGet = append(itemsForGet, items[i].RegisterPropertyToPatientForDelete...)
	//}
	return itemsForGet
}

func (items FormFills) GetRegisterPropertySetToPatient() []*FieldFill {
	itemsForGet := make([]*FieldFill, 0)
	if len(items) == 0 {
		return itemsForGet
	}
	//for i := range items {
	//	itemsForGet = append(itemsForGet, items[i].FieldFill...)
	//}
	return itemsForGet
}

func (items FormFills) GetRegisterPropertySetToPatientForDelete() []uuid.UUID {
	itemsForGet := make([]uuid.UUID, 0)
	if len(items) == 0 {
		return itemsForGet
	}
	//for i := range items {
	//	itemsForGet = append(itemsForGet, items[i].RegisterPropertySetToPatientForDelete...)
	//}
	return itemsForGet
}

func (item *FormFill) GetAggregateExistingData() string {
	res := No
	for _, FieldFill := range item.FieldFills {
		if FieldFill.GetAggregateExistingData() {
			res = Yes
			break
		}
	}
	return res
}

func (item *FormFill) Include(variantID uuid.NullUUID) string {
	res := No
	for _, FieldFill := range item.FieldFills {
		res = FieldFill.FieldFillVariantSelected(variantID)
		if res == Yes {
			break
		}
	}
	return res
}

func (item *FormFill) GetData(Field *Field) interface{} {
	// var res interface{}
	// res = No
	// for _, FieldFill := range item.FieldFills {
	// if FieldFill.FieldID == Field.ID {
	// 	res = FieldFill.GetData(Field)
	// 	break
	// }
	// }
	return nil
}

func (item *FormFill) GetScores(q *Field) int {
	sumScores := 0
	// for _, FieldFill := range item.FieldFills {
	// 	if FieldFill.FieldID == q.ID {
	// 		for _, radio := range q.FieldFillVariants {
	// 			if radio.ID == FieldFill.FieldFillVariantID {
	// 				sumScores += radio.Score
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	return sumScores
}

func (items FormFills) GetExportData(research *Form) ([][]interface{}, error) {
	results := make([][]interface{}, 0)
	for _, researchResult := range items {
		result, err := researchResult.GetXlsxData(research)
		if err != nil {
			break
		}
		results = append(results, result)
	}
	return results, nil
}

func (item *FormFill) GetXlsxData(research *Form) ([]interface{}, error) {
	result := make([]interface{}, 0)
	// result = append(result, item.Date)
	//if research.WithScores {
	//	sum := 0
	//	for _, q := range research.Fields {
	//		sum += item.GetScores(q)
	//	}
	//	results[resultN] = append(results[resultN], strconv.Itoa(sum))
	//	return err
	//}

	// variables := make(map[string]interface{})
	// for _, q := range research.Fields {
	// 	FieldFill := item.GetData(q)
	// 	for _, childField := range q.Children {
	// 		childAns := item.GetData(childField)
	// 		if childAns != nil && childAns != "" && childAns != No {
	// 			FieldFill = fmt.Sprintf("%s (%s)", FieldFill, childAns)
	// 		}
	// 	}
	// 	result = append(result, FieldFill)
	// 	variables[q.Code] = FieldFill
	// }
	// resultFormulas, err := research.Formulas.SetXlsxData(variables)
	// if err != nil {
	// 	return nil, err
	// }
	// result = append(result, resultFormulas...)
	return result, nil
}

func (item *FormFill) GetResultsMap(Fields Fields) map[string]interface{} {
	variables := make(map[string]interface{})
	for _, q := range Fields {
		FieldFill := item.GetData(q)
		variables[q.Code] = FieldFill
	}
	return variables
}

func (item *FormFill) GetAnthropometry() (uint, uint) {
	height, weight := 0, 0
	// for _, FieldFill := range item.FieldFills {
	// 	if FieldFill.Field.Code == string(AnthropomethryKeyHeight) {
	// 		height = int(FieldFill.ValueNumber)
	// 	}
	// 	if FieldFill.Field.Code == string(AnthropomethryKeyWeight) {
	// 		weight = int(FieldFill.ValueNumber)
	// 	}
	// }
	return uint(height), uint(weight)
}
