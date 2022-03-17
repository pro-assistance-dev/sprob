package paginator

import (
	"github.com/uptrace/bun"
)

func (i *Paginator) CreatePaginationQuery(query *bun.SelectQuery) {
	if i != nil {
		query = query.Limit(*i.Limit)
		query = query.Offset(*i.Offset)
	}
}
