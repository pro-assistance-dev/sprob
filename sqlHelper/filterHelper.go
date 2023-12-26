package sqlHelper

import (
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
	"github.com/uptrace/bun"
)

type QueryFilter struct {
	Col       string
	Value     string
	filter    *filter.Filter
	sorter    *sorter.Sorter
	paginator *paginator.Paginator
}

func (i *QueryFilter) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.paginator.CreatePagination(query)
	i.filter.CreateFilter(query)
	i.sorter.CreateOrder(query)
}

type FTSP struct {
	Col   string               `json:"col"`
	Value string               `json:"value"`
	F     filter.FilterModels  `json:"f"`
	S     sorter.SortModels    `json:"s"`
	P     *paginator.Paginator `json:"p"`
}

func (i *FTSP) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.P.CreatePagination(query)
	i.F.CreateFilter(query)
	i.S.CreateOrder(query)
}
