package sql

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/helpers/sql/filter"
	"github.com/pro-assistance/pro-assister/helpers/sql/paginator"
	"github.com/pro-assistance/pro-assister/helpers/sql/sorter"
	"github.com/pro-assistance/pro-assister/helpers/sql/tree"
	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type SQL struct{}

func NewSQL() *SQL {
	return &SQL{}
}

func (i *SQL) CreateQueryFilter(c *gin.Context) (*QueryFilter, error) {
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
	treeItem, err := tree.NewTreeCreator(c)
	if err != nil {
		return nil, err
	}
	return &QueryFilter{Col: col, Value: value, filter: filterItem, sorter: sorterItem, paginator: paginatorItem, treeCreator: treeItem}, nil
}

func (i *SQL) WhereLikeWithLowerTranslit(col string, search string) string {
	return fmt.Sprintf("WHERE lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g')) LIKE lower(%s)", col, "'%"+search+"%'")
}

func (i *SQL) HandleFTSPQuery(ctx context.Context, query *bun.SelectQuery) {
	i.ExtractFTSP(ctx).HandleQuery(query)
}