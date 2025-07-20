package project

import (
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun"
)

type SchemaField struct {
	bun.BaseModel `bun:"schema_fields,alias:schema_fields"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Schema   *Schema `bun:"rel:belongs-to" json:"schema"`
	SchemaID uuid.NullUUID

	Type string

	NamePascal string
	NameCamel  string
	NameCol    string
	NameRus    string
}

type SchemaFields []*SchemaField

func NewSchemaField(name string, colName string, nameRus string, t string) *SchemaField {
	return &SchemaField{
		NamePascal: name,
		NameCol:    colName,
		NameCamel:  strcase.ToLowerCamel(name),
		Type:       t,
		NameRus:    nameRus,
	}
}
