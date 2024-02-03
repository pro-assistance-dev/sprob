package tree

import (
	"fmt"

	"github.com/uptrace/bun"
)

// CreateTree method
func (item TreeModel) CreateTree(query *bun.SelectQuery) {
	query = query.NewSelect().Model(&item).ColumnExpr(item.getTableAndCols())
	fmt.Println(query.String())
}
