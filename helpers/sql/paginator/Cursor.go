package paginator

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/helpers/project"
	"github.com/pro-assistance/pro-assister/helpers/sql/filter"

	"github.com/uptrace/bun"
)

type Cursor struct {
	Operator  filter.Operator `json:"operation"`
	Column    string          `json:"column"`
	Value     string          `json:"value"`
	TableName string          `json:"tableName"`
	Model     string          `json:"model"`
	Initial   bool            `json:"initial"`
}

func (c *Cursor) createPagination(query *bun.SelectQuery) {
	if c.Initial {
		return
	}
	q := ""
	if len(c.TableName) > 0 {
		q = fmt.Sprintf("%s %s '%s'", c.getTableAndCol(), c.Operator, c.Value)
	} else {
		schema := project.SchemasLib.GetSchema(c.Model)
		q = fmt.Sprintf("%s %s '%s'", schema.GetColName(c.Column), c.Operator, c.Value)
	}
	query.Where(q)
}

func (c *Cursor) getTableAndCol() string {
	return project.SchemasLib.GetSchema(c.Model).ConcatTableCol(c.Column)
}
