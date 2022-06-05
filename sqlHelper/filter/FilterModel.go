package filter

import (
	"fmt"
	"github.com/pro-assistance/pro-assister/utilHelper"
	"github.com/uptrace/bun"
	"time"
)

// FilterModel model
type FilterModel struct {
	Table string `json:"table"`
	Col   string `json:"col"`

	Type     DataType  `json:"type,omitempty"`
	Operator Operator  `json:"operator,omitempty"`
	Date1    time.Time `json:"date1,omitempty"`
	Date2    time.Time `json:"date2,omitempty"`

	Value1  string   `json:"value1,omitempty"`
	Value2  string   `json:"value2,omitempty"`
	Boolean bool     `json:"boolean"`
	Set     []string `json:"set"`

	JoinTable      string `json:"joinTable"`
	JoinTableFK    string `json:"joinTableFK"`
	JoinTablePK    string `json:"joinTablePK"`
	JoinTableID    string `json:"joinTableId"`
	JoinTableIDCol string `json:"joinTableIdCol"`
}

// FilterModels model
type FilterModels []*FilterModel

type Operator string

const (
	Eq      Operator = "="
	Ne               = "!="
	Gt               = ">"
	Ge               = "<"
	Btw              = "between"
	Like             = "like"
	In               = "in"
	Null             = "is null"
	NotNull          = "is not null"
)

type DataType string

const (
	DateType    DataType = "date"
	NumberType           = "number"
	StringType           = "string"
	BooleanType          = "boolean"
	SetType              = "set"
	JoinType             = "join"
)

func (f *FilterModel) constructWhere(query *bun.SelectQuery) {
	q := ""
	if f.isUnary() {
		if f.Type == BooleanType {
			q = fmt.Sprintf("%s %s %t", f.getTableAndCol(), f.Operator, f.Boolean)
		} else if f.Type == DateType {
			q = fmt.Sprintf("date(%s) %s '%s'", f.getTableAndCol(), f.Operator, f.Value1)
		} else if f.isLike() {
			f.Value1 = utilHelper.NewUtilHelper("null").TranslitToRu(f.Value1)
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
	query = query.Where(q)
}

func (f *FilterModel) constructWhereIn(query *bun.SelectQuery) {
	if f.JoinTable == "" {
		query = query.Where(fmt.Sprintf("%s %s (?)", f.getTableAndCol(), f.Operator), bun.In(f.Set))
		return
	}
	q := fmt.Sprintf("EXISTS (SELECT NULL from %s where %s and %s in (?))", f.Table, f.getJoinCondition(), f.getTableAndCol())
	query = query.Where(q, bun.In(f.Set))
}

func (f *FilterModel) constructJoin(query *bun.SelectQuery) {
	if f.JoinTableID != "" {
		join := fmt.Sprintf("JOIN %s ON %s ", f.JoinTable, f.getJoinCondition())
		query = query.Join(join)
		joinTable := fmt.Sprintf("%s.%s", f.JoinTable, f.JoinTableIDCol)
		if f.Operator != In {
			query = query.Where(fmt.Sprintf("%s = ?", joinTable), f.JoinTableID)
		} else {
			query = query.Where(fmt.Sprintf("%s in (?) ", joinTable), bun.In(f.Set))
		}
		return
	}
	join := fmt.Sprintf("JOIN %s ON %s", f.JoinTable, f.getJoinCondition())
	query = query.Join(join)
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
	f.Value1 = f.Date1.Format("2006-01-02")
	if f.isBetween() {
		f.Value2 = f.Date2.Format("2006-01-02")
	}
	return
}

func (f *FilterModel) likeToString() {
	//likeOperators := map[string]string{
	//	"contains":    "%%%s%%",
	//	"notContains": "%%%s%%",
	//	"startsWith":  "%s%%",
	//	"endsWith":    "%%%s",
	//}
	f.Value1 = fmt.Sprintf("%%%s%%", f.Value1)
	return
}

func (f *FilterModel) getTableAndCol() string {
	return fmt.Sprintf("%s.%s", f.Table, f.Col)
}

func (f *FilterModel) getJoinCondition() string {
	return fmt.Sprintf("%s.%s = %s.%s", f.Table, f.JoinTablePK, f.JoinTable, f.JoinTableFK)
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

func (f *FilterModel) isSet() bool {
	return f.Operator == In
}
