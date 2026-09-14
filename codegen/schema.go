package codegen

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
	Order      Fields // порядок объявления полей (для кодогенератора)
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

// newSchema строит схему по AST-описанию структуры. typeLookup — все типы
// пакета models (структуры + алиасы слайсов) для разрешения связей.
func newSchema(structure *ast.TypeSpec, fields []*ast.Field, typeLookup map[string]*ast.TypeSpec) Schema {
	m := Schema{}
	m.SortColumn = "name"
	m.Label = "name"
	m.Value = "id"
	m.NamePascal = structure.Name.Name
	m.Fields = make(map[string]*Field)
	m.Order = make(Fields, 0)

	seen := map[string]bool{} // дедупликация по JSON-имени (встраиваемые структуры)
	for _, field := range fields {
		if isBunBaseModel(field) {
			if m.NameTable == "" {
				m.NameTable = getBunSelectTableName(parseTags(field.Tag.Value))
			}
			continue
		}
		if field.Names == nil {
			// Встраиваемая структура без тега (Relationable, Ordered, Named, StartEnd)
			// — разворачиваем её поля в текущую схему.
			for _, embedded := range getEmbeddedFields(field, typeLookup) {
				if f := makeField(embedded, typeLookup); f != nil {
					if seen[f.NameJSON] {
						continue
					}
					seen[f.NameJSON] = true
					m.Fields[f.NameCamel] = f
					m.Order = append(m.Order, f)
				}
			}
			continue
		}

		f := makeField(field, typeLookup)
		if f == nil {
			continue
		}
		if seen[f.NameJSON] {
			continue
		}
		seen[f.NameJSON] = true
		m.Fields[f.NameCamel] = f
		m.Order = append(m.Order, f)
	}
	return m
}

// makeField строит Field по ast.Field с разрешением типа и тегов.
// Возвращает nil для полей без json-тега (кроме встраиваемых — они
// обрабатываются отдельно) и для bun.BaseModel.
func makeField(field *ast.Field, typeLookup map[string]*ast.TypeSpec) *Field {
	if field.Tag == nil {
		return nil
	}
	tags := parseTags(field.Tag.Value)
	jsonName := getTagName(tags, TagJSON)
	if jsonName == "" {
		return nil
	}

	f := &Field{
		NamePascal: field.Names[0].Name,
		NameCol:    getColName(tags),
		NameCamel:  strcase.ToLowerCamel(field.Names[0].Name),
		NameJSON:   jsonName,
		TypeGo:     types.ExprString(field.Type),
		OmitEmpty:  hasOmitEmpty(tags),
	}
	if field.Doc != nil {
		if first := firstCommentLine(field.Doc.Text()); first != "" {
			f.Comment = first
		}
	}
	if jsonName == "-" {
		f.SkipJSON = true
	}

	typeGo, isSlice, elemType, isPtr, isTime, isUUID := resolveType(field.Type, typeLookup)
	f.TypeGo = typeGo
	f.IsSlice = isSlice
	f.ElemType = elemType
	f.IsPtr = isPtr
	f.IsTime = isTime
	f.IsUUID = isUUID

	// Тип для связей (InitFieldsLinksToSchemas): lowerCamel элемента.
	elem := elemType
	if !isSlice && elemType == "" {
		elem = baseTypeName(field.Type)
	}
	if elem != "" && !isTime && !isUUID {
		f.Type = strcase.ToLowerCamel(pluralize.NewClient().Singular(elem))
	}
	return f
}

// isBunBaseModel — встроенный bun.BaseModel (тег с именем таблицы).
func isBunBaseModel(field *ast.Field) bool {
	if field.Tag == nil || field.Names != nil {
		return false
	}
	return baseTypeName(field.Type) == "BaseModel"
}

// getEmbeddedFields разворачивает встраиваемую структуру в её поля
// (рекурсивно, с учётом вложенных встраиваний и bun.BaseModel).
func getEmbeddedFields(field *ast.Field, typeLookup map[string]*ast.TypeSpec) []*ast.Field {
	name := baseTypeName(field.Type)
	if name == "" {
		return nil
	}
	decl, ok := typeLookup[name]
	if !ok {
		return nil
	}
	st, ok := decl.Type.(*ast.StructType)
	if !ok {
		return nil
	}
	var result []*ast.Field
	for _, ef := range st.Fields.List {
		if isBunBaseModel(ef) {
			continue
		}
		if ef.Names == nil {
			result = append(result, getEmbeddedFields(ef, typeLookup)...)
			continue
		}
		result = append(result, ef)
	}
	return result
}

// resolveType определяет характеристики Go-типа поля.
func resolveType(expr ast.Expr, typeLookup map[string]*ast.TypeSpec) (typeGo string, isSlice bool, elemType string, isPtr bool, isTime bool, isUUID bool) {
	typeGo = types.ExprString(expr)
	switch t := expr.(type) {
	case *ast.ArrayType:
		isSlice = true
		elemType = baseTypeName(t.Elt)
		return
	case *ast.StarExpr:
		isPtr = true
		_, _, elemType, _, isTime, isUUID = resolveType(t.X, typeLookup)
		return
	case *ast.SelectorExpr:
		pkg := t.X.(*ast.Ident).Name
		switch pkg {
		case "time":
			isTime = true
		case "uuid":
			isUUID = true
		}
		return
	case *ast.Ident:
		switch t.Name {
		case "Time":
			isTime = true
			return
		case "NullUUID", "UUID":
			isUUID = true
			return
		}
		// Алиас слайса: type Rooms []*Room
		if decl, ok := typeLookup[t.Name]; ok {
			if arr, ok := decl.Type.(*ast.ArrayType); ok {
				isSlice = true
				elemType = baseTypeName(arr.Elt)
				return
			}
		}
	}
	return
}

// baseTypeName — последний сегмент имени типа (без пакета, указателя, []).
func baseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return baseTypeName(t.X)
	case *ast.ArrayType:
		return baseTypeName(t.Elt)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

func hasOmitEmpty(tags *structtag.Tags) bool {
	value, err := tags.Get(TagJSON)
	if err != nil {
		return false
	}
	for _, opt := range value.Options {
		if opt == "omitempty" {
			return true
		}
	}
	return false
}

// firstCommentLine — первая непустая строка комментария (для TS-поля).
func firstCommentLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
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
