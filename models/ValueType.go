package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ValueType struct {
	bun.BaseModel `bun:"value_types,alias:value_types"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	ValueRelation string        `bun:"type:value_type_value_relation_enum" json:"valueRelation"`
}

type ValueTypes []*ValueType

func (item *ValueType) IsString() bool {
	return item.Name == "string"
}

func (item *ValueType) IsText() bool {
	return item.Name == "text"
}

func (item *ValueType) IsNumber() bool {
	return item.Name == "number"
}

func (item *ValueType) IsDate() bool {
	return item.Name == "date"
}

func (item *ValueType) IsSet() bool {
	return item.Name == "set"
}

func (item *ValueType) IsRadio() bool {
	return item.Name == "radio"
}
