package paginator

import (
	"fmt"
	"github.com/uptrace/bun"
	"mdgkb/mdgkb-server/helpers/sqlHelper/filter"
)

type Cursor struct {
	Operator  filter.Operator `json:"operation"`
	Column    string          `json:"column"`
	Value     string          `json:"value"`
	Initial   bool            `json:"initial"`
	TableName string          `json:"tableName"`
}

func (c *Cursor) createPagination(query *bun.SelectQuery) {
	if c.Initial {
		return
	}
	q := ""
	if len(c.TableName) > 0 {
		q = fmt.Sprintf("%s.%s %s '%s'", c.TableName, c.Column, c.Operator, c.Value)
	} else {
		q = fmt.Sprintf("%s %s '%s'", c.Column, c.Operator, c.Value)
	}
	query = query.Where(q)
}
