package project

import (
	"fmt"
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun"
)

type Schema struct {
	bun.BaseModel `bun:"schemas,alias:schemas"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	NameTable  string
	NamePascal string
	NameCamel  string
	NamePlural string
	NameRus    string

	SortColumn string
	Label      string
	Value      string

	FieldsMap map[string]*SchemaField `bun:"-"  `
	Fields    SchemaFields            `bun:"rel:has-many" json:"fields"`
}

type SchemasWithCount struct {
	Schemas Schemas `json:"items"`
	Count   int     `json:"count"`
}

type (
	Schemas map[string]*Schema
)

func (items Schemas) InitFieldsLinksToSchemas() {
	for _, item := range items {
		for i := range item.FieldsMap {
			schema := items[item.FieldsMap[i].Type]
			item.FieldsMap[i].Schema = schema
		}
	}
}

const (
	TagJSON   = "json"
	TagModel  = "model"
	TagBun    = "bun"
	TagPlural = "plural"
	TagRus    = "rus"
)

func (item Schema) GetFieldsWithSchema() SchemaFields {
	fields := make(SchemaFields, 0)

	for _, field := range item.FieldsMap {
		if field.Schema == nil {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}

func (item Schema) GetFieldsCols() SchemaFields {
	fields := make(SchemaFields, 0)
	for _, field := range item.FieldsMap {
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

func (item Schema) GetField(fieldCamelCaseName string) *SchemaField {
	field := item.FieldsMap[fieldCamelCaseName]
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
	m.NameCamel = strcase.ToCamel(structure.Name.Name)
	m.FieldsMap = make(map[string]*SchemaField)

	for index, field := range fields {
		if field.Tag == nil {
			continue
		}
		tags := parseTags(field.Tag.Value)
		if index == 0 {
			m.NameTable = getBunSelectTableName(tags)
			m.NameRus = getTagName(tags, TagRus)
			// m.PluralName = ToCapCamel(m.TableName)
			continue
		}

		typeString := strcase.ToLowerCamel(pluralize.NewClient().Singular(types.ExprString(field.Type)))
		fieldSchema := NewSchemaField(field.Names[0].Name, getColName(tags), typeString, getTagName(tags, TagRus))
		m.FieldsMap[strcase.ToLowerCamel(field.Names[0].Name)] = fieldSchema
		m.Fields = append(m.Fields, fieldSchema)
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
	bunTag, err := tags.Get(TagBun)
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
