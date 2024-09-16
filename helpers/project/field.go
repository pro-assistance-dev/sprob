package project

import (
	"github.com/iancoleman/strcase"
)

type Field struct {
	Schema *Schema
	Type   string

	NamePascal string
	NameCamel  string
	NameCol    string
}

type Fields []*Field

func NewField(name string, colName string, t string) *Field {
	return &Field{
		NamePascal: name,
		NameCol:    colName,
		NameCamel:  strcase.ToLowerCamel(name),
		Type:       t,
	}
}
