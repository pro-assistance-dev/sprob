package f

import (
	"fmt"
	"time"

	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

// FilterModel model
type Where struct {
	Field    string   `json:"field"`
	Operator Operator `json:"operator,omitempty"`
	Value    any      `json:"value,omitempty"`

	expr string
}

type Wheres []*Where

func (item Wheres) Construct(table string, schema project.Schema, bunF func(string, ...interface{}) *bun.SelectQuery) {
	for _, where := range item {
		where.Construct(table, schema)
		bunF(where.expr, where.Value)
	}
}

func (item *Where) isNull() bool {
	return item.Operator == Null || item.Operator == NotNull
}

func (item *Where) isUnary() bool {
	return item.Operator == Eq || item.Operator == Ne || item.Operator == Gt || item.Operator == Ge || item.Operator == Like
}

func (item *Where) Where(q *bun.SelectQuery) {
	if item.isUnary() {
		q.Where(item.expr, item.Value)
	}
	if item.Operator == In {
		q.Where(item.expr, bun.In(item.Value))
	}
}

func (item *Where) Construct(table string, schema project.Schema) {
	field := schema.GetField(item.Field)
	if item.isUnary() {
		if field.Type == "time.Time" {
			item.Value = item.Value.(time.Time).Format("2006-01-02 15:04:05")
		}
		item.expr = item.whereBool(table, field.NameCol)
	}

	if item.isNull() {
		item.expr = item.whereNull(table, field.NameCol)
	}

	if item.Operator == In {
		item.expr = item.whereIn(table, field.NameCol)
	}
}

func (item *Where) whereBool(table string, field string) string {
	return fmt.Sprintf("%s.%s %s ?", table, field, item.Operator)
}

// func (item *Where) whereTime(table string, field string, value any) string {
// 	return fmt.Sprintf("%s.%s = ?", table, field)
// }

func (item *Where) whereIn(table string, field string) string {
	return fmt.Sprintf("%s.%s in (?))", table, field)
}

func (item *Where) whereNull(table string, field string) string {
	return fmt.Sprintf("%s.%s = ?", table, field)
}
