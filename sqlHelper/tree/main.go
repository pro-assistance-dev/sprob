package tree

import (
	"fmt"

	"github.com/uptrace/bun"
)

// CreateTree method
func (item TreeModel) CreateTree(query *bun.SelectQuery) {
	query = query.NewSelect().Model(&item).ColumnExpr(fmt.Sprintf("%s", item.getTableAndCols()))
}
