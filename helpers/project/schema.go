package project

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

type Schema struct {
	NameTable  string
	NamePascal string
	NameCamel  string
	NamePlural string

	SortColumn string
	Label      string
	Value      string
	Fields     map[string]*Field
}

type (
	Schemas map[string]*Schema
)

func (items Schemas) InitFieldsLinksToSchemas() {
	for _, item := range items {
		for i := range item.Fields {
			schema := items[item.Fields[i].Type]
			item.Fields[i].Schema = schema
		}
	}
}

const (
	TagJSON   = "json"
	TagModel  = "model"
	TagBun    = "bun"
	TagPlural = "plural"
)

func (item Schema) GetFieldsWithSchema() Fields {
	fields := make(Fields, 0)
	for _, field := range item.Fields {
		if field.Schema == nil {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func (item Schema) GetFieldsCols() Fields {
	fields := make(Fields, 0)
	for _, field := range item.Fields {
		if field.Schema != nil {
			continue
		}
		fields = append(fields, field)
	}
	return fields
}

func (item *Schema) ConcatTableCols() []string {
	cols := make([]string, 0)
	if item == nil {
		return cols
	}
	for _, field := range item.GetFieldsCols() {
		cols = append(cols, item.ConcatTableCol(field.NameCamel))
	}
	return cols
}

func (item Schema) ConcatTableCol(colNameInCamelCase string) string {
	return fmt.Sprintf("%s.%s", item.NameTable, item.GetColName(colNameInCamelCase))
}

func (items Schemas) GetSchema(schemaName string) *Schema {
	return items[schemaName]
}

func (item Schema) GetTableName() string {
	return item.NameTable
}

func (item Schema) GetField(fieldCamelCaseName string) *Field {
	field := item.Fields[fieldCamelCaseName]
	return field
}

func (item Schema) GetColName(colNameInCamelCase string) string {
	return item.GetField(colNameInCamelCase).NameCol
}

func newSchema(structure *ast.TypeSpec, fields []*ast.Field) Schema {
	m := Schema{}
	m.SortColumn = "name"
	m.Label = "name"
	m.Value = "id"
	// m.Key = strcase.ToLowerCamel(structure.Name.Name)
	m.NamePascal = structure.Name.Name
	m.Fields = make(map[string]*Field)

	for index, field := range fields {
		if field.Tag == nil {
			continue
		}
		tags := parseTags(field.Tag.Value)
		if index == 0 {
			m.NameTable = getBunSelectTableName(tags)
			// m.PluralName = ToCapCamel(m.TableName)
			continue
		}

		typeString := strcase.ToLowerCamel(pluralize.NewClient().Singular(types.ExprString(field.Type)))
		m.Fields[strcase.ToLowerCamel(field.Names[0].Name)] = NewField(field.Names[0].Name, getColName(tags), typeString)
	}
	return m
}

func getTagName(tags *structtag.Tags, tag string) string {
	value, err := tags.Get(tag)
	if err != nil {
		return ""
	}
	return value.Name
}

func getColName(tags *structtag.Tags) string {
	bunTag := getTagName(tags, TagBun)
	if bunTag != "-" && bunTag != "" && !strings.Contains(bunTag, ":") {
		return bunTag
	}
	return strcase.ToSnake(getTagName(tags, TagJSON))
}

func getBunSelectTableName(tags *structtag.Tags) string {
	bunTag, err := tags.Get("bun")
	if err != nil {
		return ""
	}
	tableName := bunTag.Name
	for _, opt := range bunTag.Options {
		parts := strings.Split(opt, ":")
		if len(parts) == 2 && parts[0] == "select" {
			tableName = parts[1]
		}
	}
	return tableName
}

func parseTags(tagString string) *structtag.Tags {
	tag, err := strconv.Unquote(tagString)
	if err != nil {
		panic(err)
	}
	tags, err := structtag.Parse(tag)
	if err != nil {
		panic(fmt.Sprintf("%s: %s", err.Error(), tagString))
	}
	return tags
}
