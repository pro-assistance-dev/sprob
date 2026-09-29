package sorter

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
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
			// ⚠️ Без fmt.Println (загрязнял прод-лог). Колонка — через `?`→`bun.Ident`
			// (резолвлена схемой), порядок — из белого списка `orderFrag`, поэтому
			// значение из запроса не может стать произвольным SQL-фрагментом.
			query = query.OrderExpr("? ?", bun.Ident(sort.getTableAndCol()), schema.SafeQuery(orderFrag(sort.Order), nil))
		}
		return
	}
	for _, sort := range defaultSort {
		query = query.Order(sort)
	}
}

// CreateOrder method
func (items SortModels) CreateOrder(query *bun.SelectQuery, defaultSort ...string) {
	if len(items) != 0 {
		for _, sort := range items {
			if sort == nil {
				sort.Order = Asc
			}
			query = query.OrderExpr("? ?", bun.Ident(sort.getTableAndCol()), schema.SafeQuery(orderFrag(sort.Order), nil))
		}
		return
	}
	for _, sort := range defaultSort {
		query = query.Order(sort)
	}
}

// orderFrag — только `asc`/`desc`; любое иное значение сводим к `asc`,
// чтобы значение из запроса не могло стать произвольным SQL-фрагментом.
func orderFrag(o Orders) string {
	if o == Desc {
		return string(Desc)
	}
	return string(Asc)
}
