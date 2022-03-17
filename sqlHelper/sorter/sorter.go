package sorter

import (
	"fmt"
	"github.com/uptrace/bun"
)

type Sorter struct {
	ID         *string
	sortModels SortModels
}

// CreateOrder method
func (i *Sorter) CreateOrder(query *bun.SelectQuery, defaultSort ...string) {
	if len(i.sortModels) != 0 {
		for _, sort := range i.sortModels {
			if sort == nil {
				sort.Order = Asc
			}
			query = query.OrderExpr(fmt.Sprintf("%s %s", sort.getTableAndCol(), sort.Order))
		}
		return
	}
	for _, sort := range defaultSort {
		query = query.Order(sort)
	}
}
