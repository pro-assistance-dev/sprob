package project

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
)

type Schema struct {
	Name       string
	SortColumn string
	Label      string
	Value      string
	Key        string
	TableName  string
	PluralName string
	Fields     map[string]*Field
}

type (
	Schemas map[string]*Schema
)

func (items Schemas) InitFieldsLinksToSchemas() {
	for _, item := range items {
		for i := range item.Fields {
			item.Fields[i].Schema = items[item.Fields[i].Type]
		}
	}
}

const (
	TagJSON   = "json"
	TagModel  = "model"
	TagBun    = "bun"
	TagPlural = "plural"
)

func (items Schemas) GetSchema(schemaName string) *Schema {
	return items[schemaName]
}

func (item Schema) GetTableName() string {
	return item.TableName
}

func (item Schema) GetColName(colNameInCamelCase string) string {
	field := item.Fields[colNameInCamelCase]
	name := field.ColName
	return name
}

func newSchema(structure *ast.TypeSpec, fields []*ast.Field) Schema {
	m := Schema{}
	m.SortColumn = "name"
	m.Label = "name"
	m.Value = "id"
	m.Key = ToLowerCamel(structure.Name.Name)
	m.Name = structure.Name.Name

	for index, field := range fields {
		if field.Tag == nil {
			continue
		}
		tags := parseTags(field.Tag.Value)
		if index == 0 {
			m.TableName = getBunSelectTableName(tags)
			m.PluralName = ToCapCamel(m.TableName)
			continue
		}
		m.Fields[field.Names[0].Name] = NewField(field.Names[0].Name, getColName(tags))
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
	return toSnake(getTagName(tags, TagJSON))
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
