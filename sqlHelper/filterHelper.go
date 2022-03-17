package sqlHelper

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
)

type QueryFilter struct {
	ID        *string
	Filter    *filter.Filter
	Sorter    *sorter.Sorter
	Paginator *paginator.Paginator
}

func (i *SQLHelper) CreateQueryFilter(c *gin.Context) (*QueryFilter, error) {
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
	id := c.Param("id")
	return &QueryFilter{ID: &id, Filter: filterItem, Sorter: sorterItem, Paginator: paginatorItem}, nil
}
