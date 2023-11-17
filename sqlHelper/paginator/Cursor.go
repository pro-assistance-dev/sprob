package paginator

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/uptrace/bun"
)

type Cursor struct {
	Operator  filter.Operator `json:"operation"`
	Column    string          `json:"column"`
	Value     string          `json:"value"`
	Initial   bool            `json:"initial"`
	TableName string          `json:"tableName"`
	Model     string          `json:"model"`
}

func (c *Cursor) createPagination(query *bun.SelectQuery) {
	if c.Initial {
		return
	}
	q := ""
	if len(c.TableName) > 0 {
		q = fmt.Sprintf("%s %s '%s'", c.getTableAndCol(), c.Operator, c.Value)
	} else {
		schema := projecthelper.SchemasLib.GetSchema(c.Model)
		q = fmt.Sprintf("%s %s '%s'", schema.GetCol(c.Column), c.Operator, c.Value)
	}
	query = query.Where(q)
}

func (c *Cursor) getTableAndCol() string {
		schema := projecthelper.SchemasLib.GetSchema(c.Model)
		return fmt.Sprintf("%s.%s", schema.GetTableName(), schema.GetCol(c.Column))
}
