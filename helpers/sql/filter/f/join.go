package f

import (
	"fmt"

	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

// FilterModel model
type Join struct {
	Model string `json:"model"`
	Field string `json:"field"`
	Value any    `json:"value,omitempty"`

	Wheres Wheres `json:"wheres"`
}

type Joins []*Join

func (joins Joins) Construct(query *bun.SelectQuery, table string) {
	for _, join := range joins {
		join.construct(query, table)
	}
}

func (join *Join) construct(query *bun.SelectQuery, table string) {
	joinSchema := project.SchemasLib.GetSchema(join.Model)
	joinTable := joinSchema.GetTableName()
	joinField := joinSchema.GetField(join.Field)

	joinCondition := fmt.Sprintf("%s.id = %s.%s", table, joinTable, joinField.NameCol)
	query.Join(joinCondition)
	join.Wheres.Construct(joinTable, *joinSchema, query.JoinOn)
}
