package codegen

import (
	"github.com/iancoleman/strcase"
)

type Field struct {
	Schema *Schema
	Type   string

	NamePascal string
	NameCamel  string
	NameCol    string

	// --- В1 (генерация TS-классов): метаданные поля для кодогенератора ---
	NameJSON  string // имя в JSON (из json-тега, без omitempty)
	TypeGo    string // исходный Go-тип (types.ExprString)
	ElemType  string // для слайсов — тип элемента (Pascal), иначе ""
	IsSlice   bool   // слайс: []*X / []X / алиас Xxxs []*X
	IsPtr     bool   // указатель *X
	IsTime    bool   // time.Time / *time.Time
	IsUUID    bool   // uuid.NullUUID
	OmitEmpty bool   // json:"...,omitempty"
	SkipJSON  bool   // json:"-" — не сериализуется в JSON
	Comment   string // первая строка Go-комментария над полем (для TS)
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
