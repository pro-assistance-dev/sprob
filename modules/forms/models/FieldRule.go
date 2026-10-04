package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// FieldRule — условие показа вопроса («если, то»): вопрос (FieldID) показывается
// только когда ответ на другой вопрос (DependsOnFieldID) удовлетворяет условию
// (Operator + Value). Несколько правил у поля соединяются через AND — вопрос
// виден, когда выполнены ВСЕ. Это минимальная, предсказуемая модель ветвления
// (полноценный конструктор выражений — отдельная задача).
const (
	// OperatorEquals — ответ равен значению.
	OperatorEquals = "equals"
	// OperatorNotEquals — ответ не равен значению.
	OperatorNotEquals = "notEquals"
	// OperatorContains — среди выбранных вариантов есть значение.
	OperatorContains = "contains"
	// OperatorNotContains — среди выбранных вариантов нет значения.
	OperatorNotContains = "notContains"
	// OperatorAnswered — на вопрос ответили (любой ответ).
	OperatorAnswered = "answered"
	// OperatorNotAnswered — на вопрос не ответили.
	OperatorNotAnswered = "notAnswered"
	// OperatorGreater — числовой ответ больше значения.
	OperatorGreater = "greater"
	// OperatorLess — числовой ответ меньше значения.
	OperatorLess = "less"
)

// FieldRules — правила показа вопроса.
type FieldRules []*FieldRule

type FieldRule struct {
	bun.BaseModel `bun:"field_rules,alias:field_rules"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	// FieldID — вопрос, к которому относится правило (он показывается/скрывается).
	FieldID uuid.NullUUID `bun:"type:uuid" json:"fieldId"`

	// DependsOnFieldID — вопрос-условие (чьи ответы проверяются).
	DependsOnFieldID uuid.NullUUID `bun:"type:uuid" json:"dependsOnFieldId"`
	// DependsOnFieldCode — код вопроса-условия (для рантайма без полного дерева).
	DependsOnFieldCode string `json:"dependsOnFieldCode"`

	// Operator — один из Operator* (equals/contains/greater/...).
	Operator string `json:"operator"`
	// Value — сравниваемое значение (id варианта ответа, текст, число).
	Value string `json:"value"`

	Order uint `bun:"item_order" json:"order"`
}

// SetIDForChildren проставляет правилу id вопроса.
func (item *FieldRule) SetIDForChildren(fieldID uuid.NullUUID) {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.FieldID = fieldID
}

func (items FieldRules) SetIDForChildren(fieldID uuid.NullUUID) {
	for i := range items {
		items[i].SetIDForChildren(fieldID)
	}
}
