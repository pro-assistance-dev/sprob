package projecthelper

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/fatih/structtag"
)

type Schema map[string]string
type Schemas map[string]Schema

func (items Schemas) GetSchema(schemaName string) Schema {
	return items[schemaName]
}

func (item Schema) GetCol(colNameInCamelCase string) string {
	fmt.Println(item[colNameInCamelCase])
	return item[colNameInCamelCase]
}

func (item Schema) GetTableName() string {
	return item["tableName"]
}

func getSchema(structure *ast.TypeSpec, fields []*ast.Field) Schema {
	m := Schema{}
	m["sortColumn"] = "name"
	m["label"] = "name"
	m["value"] = "id"
	for index, field := range fields {
		if field.Tag == nil {
			continue
		}
		tags := parseTags(field.Tag.Value)
		if index == 0 {
			m["tableName"] = getBunSelectTableName(tags)
			continue
		}
		m[getJSONName(tags)] = getColName(tags)
	}
	m["key"] = ToLowerCamel(structure.Name.Name)
	return m
}

func getJSONName(tags *structtag.Tags) string {
	jsonName, err := tags.Get("json")
	if err != nil {
		return ""
	}
	return jsonName.Name
}

func getColName(tags *structtag.Tags) string {
	bunTag, err := tags.Get("bun")
	if err == nil && bunTag.Name != "-" && bunTag.Name != "" && !strings.Contains(bunTag.Name, ":") {
		return bunTag.Name
	}
	return toSnake(getJSONName(tags))
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
