package t

import (
	"fmt"

	"github.com/uptrace/bun"
)

// CreateTree method
func (item treeModel) CreateTree(query *bun.SelectQuery, cols ...string) {
	query = query.OrderExpr(fmt.Sprintf("%s", item.getTableAndCols()))
}
