package sqlHelper

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
	"github.com/uptrace/bun"
)

type QueryFilter struct {
	col       string
	value     string
	filter    *filter.Filter
	sorter    *sorter.Sorter
	paginator *paginator.Paginator
}

func (i *SQLHelper) CreateQueryFilter(c *gin.Context) (*QueryFilter, error) {
	col := c.Query("col")
	value := c.Query("value")
	filterItem, err := filter.NewFilter(c)
	if err != nil {
		return nil, err
	}
	sorterItem, err := sorter.NewSorter(c)
	if err != nil {
		return nil, err
	}
	paginatorItem, err := paginator.NewPaginator(c)
	if err != nil {
		return nil, err
	}
	return &QueryFilter{col: col, value: value, filter: filterItem, sorter: sorterItem, paginator: paginatorItem}, nil
}

func (i *QueryFilter) HandleQuery(query *bun.SelectQuery) {
	i.paginator.CreatePagination(query)
	i.filter.CreateFilter(query)
	i.sorter.CreateOrder(query)
}
