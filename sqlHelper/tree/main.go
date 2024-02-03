package tree

import (
	"fmt"

	"github.com/uptrace/bun"
)

// CreateTree method
func (item TreeModel) CreateTree(query *bun.SelectQuery, cols ...string) {
	query = query.OrderExpr(fmt.Sprintf("%s", item.getTableAndCols()))
}
