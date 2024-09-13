package project

type Field struct {
	Schema  *Schema
	Name    string
	ColName string
	Type    string
}

func NewField(name string, colName string) *Field {
	return &Field{Name: name, ColName: colName}
}
