package project

import (
	"fmt"
	"go/ast"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
)

type (
	Schema  map[string]string
	Schemas map[string]Schema
)

const (
	TagJSON   = "json"
	TagModel  = "model"
	TagBun    = "bun"
	TagPlural = "plural"
)

func (items Schemas) GetSchema(schemaName string) Schema {
	return items[schemaName]
}

func (item Schema) GetCol(colNameInCamelCase string) string {
	return item[colNameInCamelCase]
}

var modelFields = []string{"sortColumn", "label", "value", "structName", "tableName", "plural"}

func (item Schema) GetFields() []string {
	fields := make([]string, 0)
	for v := range item {
		if slices.Contains(fields, v) {
			continue
		}
		fields = append(fields, v)
	}
	return fields
}

func (item Schema) GetTableName() string {
	return item["tableName"]
}

func getSchema(structure *ast.TypeSpec, fields []*ast.Field) Schema {
	m := Schema{}
	m["sortColumn"] = "name"
	m["label"] = "name"
	m["value"] = "id"
	m["key"] = ToLowerCamel(structure.Name.Name)
	m["structName"] = structure.Name.Name

	for index, field := range fields {
		if field.Tag == nil {
			continue
		}
		tags := parseTags(field.Tag.Value)
		if index == 0 {
			m["tableName"] = getBunSelectTableName(tags)
			m["plural"] = ToCapCamel(m["tableName"])
			continue
		}
		m[getTagName(tags, TagJSON)] = getColName(tags)
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
