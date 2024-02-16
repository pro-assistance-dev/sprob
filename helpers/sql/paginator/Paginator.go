package paginator

import (
	"github.com/uptrace/bun"
)

func (i *Paginator) CreatePaginationQuery(query *bun.SelectQuery) {
	if i != nil {
		query.Limit(*i.Limit)
		query.Offset(*i.Offset)
	}
}
