package filter

import (
	"fmt"
	"time"

	project "github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/util"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// FilterModel model
type FilterModel struct { //nolint:golint
	ID     string `json:"id"`
	Table  string `json:"table"`
	Col    string `json:"col"`
	Model  string `json:"model"`
	Value1 string `json:"value1,omitempty"`

	Type     DataType  `json:"type,omitempty"`
	Operator Operator  `json:"operator,omitempty"`
	Date1    time.Time `json:"date1,omitempty"`
	Date2    time.Time `json:"date2,omitempty"`

	Value2  string   `json:"value2,omitempty"`
	Set     []string `json:"set"`
	Boolean bool     `json:"boolean"`

	JoinTable      string `json:"joinTable"`
	JoinTableModel string `json:"joinTableModel"`
	JoinTableFK    string `json:"joinTableFK"`
	JoinTablePK    string `json:"joinTablePK"`
	JoinTableID    string `json:"joinTableId"`
	JoinTableIDCol string `json:"joinTableIdCol"`

	JoinIn   []string
	JoinSets [][]string

	ignore bool
}

// FilterModels model
type FilterModels []*FilterModel //nolint:golint

type Operator string

const (
	Eq      Operator = "="
	Ne      Operator = "!="
	Gt      Operator = ">"
	Ge      Operator = "<"
	Btw     Operator = "between"
	Like    Operator = "like"
	In      Operator = "in"
	Null    Operator = "is null"
	NotNull Operator = "is not null"
)

type DataType string

const (
	DateType    DataType = "date"
	NumberType  DataType = "number"
	StringType  DataType = "string"
	BooleanType DataType = "boolean"
	SetType     DataType = "set"
	JoinType    DataType = "join"
)

// TranslitToRu — транслитерация латиницы в кириллицу для LIKE-поиска
// (пользователь может набрать не в той раскладке). Вынесена в переменную,
// чтобы тесты могли подменить без БД/утилит хелпера.
var TranslitToRu = func(s string) string { return util.NewUtil("null").TranslitToRu(s) }

// opFrag — фрагмент SQL для оператора сравнения.
//
// ⚠️ Оператор НЕЛЬЗЯ отдавать ни параметром, ни подстановкой строки:
// `?` в качестве оператора даёт `col = ?` с ушедшим значением, а конкатенация —
// инъекцию. Берём его через `schema.SafeQuery`: сюда попадают только значения
// из нашего же `const`-набора `Operator`, то есть строка безопасна по построению.
func opFrag(op Operator) schema.QueryWithArgs {
	return schema.SafeQuery(string(op), nil)
}

// constructWhere добавляет условие фильтра ПАРАМЕТРАМИ (`?`), а не конкатенацией.
//
// ⚠️ Колонка подставляется как `bun.Ident` и только после резолва через схему
// модели (`getTableAndCol`) — произвольное имя из запроса до SQL не дойдёт.
// Значение — ВСЕГДА параметр: раньше `fmt.Sprintf("... '%s'", f.Value1)` давало
// SQL-инъекцию через значение (апостроф ломал запрос), т.к. `Value1` приходит
// от клиента.
func (f *FilterModel) constructWhere(query *bun.SelectQuery) {
	col := f.getTableAndCol()

	if f.isUnary() {
		switch f.Type {
		case BooleanType:
			query.Where("? ? ?", bun.Ident(col), opFrag(f.Operator), f.Boolean)
		case DateType:
			query.Where("? ? ?", bun.Ident(col), opFrag(f.Operator), f.Value1)
		default:
			if f.isLike() {
				f.Value1 = TranslitToRu(f.Value1)
				f.likeToString()
				// Регистронезависимо и без пунктуации — это ВЫРАЖЕНИЕ колонки
				// (не имя), поэтому `?` его не отквотит — подставляем через
				// UnsafeIdent, но только по резолвленной схеме колонке.
				expr := fmt.Sprintf("lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g'))", col)
				query.Where("? ? lower(?)", schema.UnsafeIdent(expr), opFrag(f.Operator), f.Value1)
			} else {
				query.Where("? ? ?", bun.Ident(col), opFrag(f.Operator), f.Value1)
			}
		}
	}
	if f.isBetween() {
		query.Where("? ? ? and ?", bun.Ident(col), opFrag(f.Operator), f.Value1, f.Value2)
	}
	if f.isNull() {
		query.Where("? ?", bun.Ident(col), opFrag(f.Operator))
	}
}

func (f *FilterModel) constructWhereIn(query *bun.SelectQuery) {
	if len(f.Set) == 0 {
		return
	}
	if f.Type != JoinType {
		query.Where("? ? (?)", bun.Ident(f.getTableAndCol()), opFrag(f.Operator), bun.In(f.Set))
		return
	}
	// ⚠️ Join-ветка: и таблица, и колонка резолвятся через схемы (имя в SQL),
	// а значения (`Set`) уходят параметром.
	if len(f.Set) > 1 {
		query.Where("?", schema.UnsafeIdent(fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in ?)", f.Table, f.getJoinCondition(), f.getTableAndCol())), bun.In(f.Set))
		return
	}
	query.Where("?", schema.UnsafeIdent(fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in (?))", f.Table, f.getJoinCondition(), f.getTableAndCol())), bun.In(f.Set))
}

func (f *FilterModel) constructJoin(query *bun.SelectQuery) {
	f.constructJoinV3(query)
	// if f.JoinTableID != "" && f.Version != "v2" {
	// 	join := fmt.Sprintf("JOIN %s ON %s ", f.JoinTable, f.getJoinCondition())
	// 	query.Join(join)
	// 	joinTable := fmt.Sprintf("%s.%s", f.JoinTable, f.JoinTableIDCol)
	// 	if f.Operator != In {
	// 		query.Where("? = ?", bun.Ident(joinTable), f.JoinTableID)
	// 	} else {
	// 		query.Where("? in (?)", bun.Ident(joinTable), bun.In(f.Set))
	// 	}
	// 	return
	// }
	// joinTable := f.JoinTable
	// joinModel := project.Schema{}
	// if f.Version == "v2" {
	// 	joinModel = project.SchemasLib.GetSchema(f.JoinTableModel)
	// 	joinTable = joinModel.GetTableName()
	// }
	// join := fmt.Sprintf("JOIN %s ON %s", joinTable, f.getJoinCondition())
	// query.Join(join)
	// if f.JoinTableID != "" {
	// 	if f.Operator != In {
	// 		query.Where("? = ?", bun.Ident(joinModel.GetCol(f.JoinTableIDCol)), f.JoinTableID)
	// 	} else {
	// 		query.Where("? in (?)", bun.Ident(joinTable), bun.In(f.Set))
	// 	}
	// }
}

func (f *FilterModel) constructJoinV3(query *bun.SelectQuery) {
	model := project.SchemasLib.GetSchema(f.Model)
	joinModel := project.SchemasLib.GetSchema(f.JoinTableModel)

	modelTable := model.GetTableName()
	joinTable := joinModel.GetTableName()
	joinCondition := fmt.Sprintf("%s.id = %s.%s", modelTable, joinTable, joinModel.GetColName(f.Model+"Id"))

	query.Join(fmt.Sprintf("JOIN %s", joinTable))
	query.JoinOn(joinCondition)
	// joinTable := joinModel.GetTableName()

	for _, joinIn := range f.JoinIn {
		joinStr := "%s.%s in (?)"
		if len(f.JoinSets) == 1 {
			joinStr = "%s.%s in ?"
		}
		joinCondition = fmt.Sprintf(joinStr, joinTable, joinModel.GetColName(joinIn))
		query.JoinOn(joinCondition, bun.In(f.JoinSets))
	}

	// if f.Operator == In {
	// 	col := joinModel.GetColName(f.Col)
	// 	modelTable := model.GetTableName()
	// 	joinTable := joinModel.GetTableName()
	// 	joinCondition := fmt.Sprintf("%s.id = %s.%s", modelTable, joinTable, joinModel.GetColName(f.Model+"Id"))
	// 	q := fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in (?))", joinModel.GetTableName(), joinCondition, col)
	// 	query.Where(q, bun.In(f.Set))
	// 	// query.Where("?.? in (?)", bun.Ident(joinModel.GetTableName()), bun.Ident(col), bun.In(f.Set))
	// } else {
	// }
}

// func constructTextWhere(tbl *bun.SelectQuery, field string, operator string, options ...models.filter) *bun.SelectQuery {
//	operators := map[string]string{
//		"equals":      "%s = ?",
//		"notEqual":    "%s <> ?",
//		"contains":    "%s LIKE ?",
//		"notContains": "%s NOT LIKE ?",
//		"startsWith":  "%s LIKE ?",
//		"endsWith":    "%s LIKE ?",
//	}
//	if operator == "" {
//		tbl = constructQuery(tbl, operators[*options[0].Type], "", field, operator, likeMix(*options[0].Type, fmt.Sprintf("%v", *options[0].filter)))
//	} else {
//		tbl = constructQuery(tbl, operators[*options[0].Type], operators[*options[1].Type], field, operator, likeMix(*options[0].Type, (*options[0].filter).(string)), likeMix(*options[1].Type, (*options[1].filter).(string)))
//	}
//	return tbl
//}
//
//func likeMix(typeOperator, filter string) (result string) {
//	likeOperators := map[string]string{
//		"contains":    "%%%s%%",
//		"notContains": "%%%s%%",
//		"startsWith":  "%s%%",
//		"endsWith":    "%%%s",
//	}
//	likePhrase, ok := likeOperators[typeOperator]
//	if ok {
//		return fmt.Sprintf(likePhrase, filter)
//	}
//	return filter
//}
//
//func getTimeString(date string) string {
//	resultDate := time.Now()
//	resultDate, err := time.Parse("2006-01-02 15:04:05", date)
//	if err != nil {
//		resultDate = time.Date(resultDate.Year(), resultDate.Month(), resultDate.Day(), 0, 0, 0, 0, time.UTC)
//	}
//	location, _ := time.LoadLocation("Europe/Moscow")
//	result := resultDate.In(location).Format("2006-01-02")
//	return result
//}
//

func (f *FilterModel) datesToString() {
	f.Value1 = f.Date1.Format("2006-01-02 15:04:05")
	if f.isBetween() {
		f.Value2 = f.Date2.Format("2006-01-02 15:04:05")
	}
}

func (f *FilterModel) likeToString() {
	//likeOperators := map[string]string{
	//	"contains":    "%%%s%%",
	//	"notContains": "%%%s%%",
	//	"startsWith":  "%s%%",
	//	"endsWith":    "%%%s",
	//}
	f.Value1 = fmt.Sprintf("%%%s%%", f.Value1)
}

func (f *FilterModel) getTableAndCol() string {
	return project.SchemasLib.GetSchema(f.Model).ConcatTableCol(f.Col)
}

func (f *FilterModel) getJoinCondition() string {
	model := project.SchemasLib.GetSchema(f.Model)
	joinModel := project.SchemasLib.GetSchema(f.JoinTableModel)
	return fmt.Sprintf("%s = %s", model.ConcatTableCol(f.JoinTablePK), joinModel.ConcatTableCol(f.JoinTableFK))
}

//	func (f *FilterModel) getJoinExpression(model *project.Schema, joinModel *project.Schema) string {
//		modelTable := model.GetTableName()
//		joinTable := joinModel.GetTableName()
//		joinCondition := fmt.Sprintf("%s.id = %s.%s", modelTable, joinTable, joinModel.GetColName(f.Model+"Id"))
//		return fmt.Sprintf("JOIN %s ON %s", joinTable, joinCondition)
//	}
func (f *FilterModel) isUnary() bool {
	return f.Operator == Eq || f.Operator == Ne || f.Operator == Gt || f.Operator == Ge || f.Operator == Like
}

func (f *FilterModel) isLike() bool {
	return f.Operator == Like
}

func (f *FilterModel) isBetween() bool {
	return f.Operator == Btw
}

func (f *FilterModel) isNull() bool {
	return f.Operator == Null || f.Operator == NotNull
}

// func (f *FilterModel) isSet() bool {
// 	return f.Operator == In
// }
