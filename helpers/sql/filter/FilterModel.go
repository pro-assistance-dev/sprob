package filter

import (
	"fmt"
	"time"

	"github.com/pro-assistance/pro-assister/helpers/project"
	"github.com/pro-assistance/pro-assister/helpers/util"

	"github.com/uptrace/bun"
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

func (f *FilterModel) constructWhere(query *bun.SelectQuery) {
	q := ""
	if f.isUnary() {
		if f.Type == BooleanType {
			q = fmt.Sprintf("%s %s %t", f.getTableAndCol(), f.Operator, f.Boolean)
		} else if f.Type == DateType {
			q = fmt.Sprintf("%s %s '%s'", f.getTableAndCol(), f.Operator, f.Value1)
		} else if f.isLike() {
			f.Value1 = util.NewUtil("null").TranslitToRu(f.Value1)
			f.likeToString()
			col := fmt.Sprintf("lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g'))", f.getTableAndCol())
			q = fmt.Sprintf("%s %s lower('%s')", col, f.Operator, f.Value1)
		} else {
			q = fmt.Sprintf("%s %s '%s'", f.getTableAndCol(), f.Operator, f.Value1)
		}
	}
	if f.isBetween() {
		q = fmt.Sprintf("%s %s '%s' and '%s'", f.getTableAndCol(), f.Operator, f.Value1, f.Value2)
	}
	if f.isNull() {
		q = fmt.Sprintf("%s %s", f.getTableAndCol(), f.Operator)
	}
	query.Where(q)
}

func (f *FilterModel) constructWhereIn(query *bun.SelectQuery) {
	// if f.JoinTable == "" {
	// 	query.Where(fmt.Sprintf("%s %s (?)", f.getTableAndCol(), f.Operator), bun.In(f.Set))
	// 	return
	// }
	q := fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in (?))", f.Table, f.getJoinCondition(), f.getTableAndCol())
	query.Where(q, bun.In(f.Set))
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
	query.Join(f.getJoinExpression(model, joinModel))
	if f.Operator == In {
		col := joinModel.GetCol(f.Col)
		q := fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in (?))", joinModel.GetTableName(), f.getJoinExpression(model, joinModel), col)
		query.Where(q, bun.In(f.Set))
		// query.Where("?.? in (?)", bun.Ident(joinModel.GetTableName()), bun.Ident(col), bun.In(f.Set))
	}
}

//
//func constructTextWhere(tbl *bun.SelectQuery, field string, operator string, options ...models.filter) *bun.SelectQuery {
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
	schema := project.SchemasLib.GetSchema(f.Model)
	return fmt.Sprintf("%s.%s", schema.GetTableName(), schema.GetCol(f.Col))
}

func (f *FilterModel) getJoinCondition() string {
	model := project.SchemasLib.GetSchema(f.Model)
	joinModel := project.SchemasLib.GetSchema(f.JoinTableModel)
	return fmt.Sprintf("%s.%s = %s.%s", model.GetTableName(), model.GetCol(f.JoinTablePK), joinModel.GetTableName(), joinModel.GetCol(f.JoinTableFK))
}

func (f *FilterModel) getJoinExpression(model project.Schema, joinModel project.Schema) string {
	modelTable := model.GetTableName()
	joinTable := joinModel.GetTableName()
	joinCondition := fmt.Sprintf("%s.id = %s.%s", modelTable, joinTable, joinModel.GetCol(f.Model+"Id"))
	return fmt.Sprintf("JOIN %s ON %s", joinTable, joinCondition)
}

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
