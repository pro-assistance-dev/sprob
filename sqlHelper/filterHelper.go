package sqlHelper

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
	"github.com/pro-assistance/pro-assister/sqlHelper/tree"
	"github.com/uptrace/bun"
)

type QueryFilter struct {
	Col         string
	Value       string
	filter      *filter.Filter
	sorter      *sorter.Sorter
	paginator   *paginator.Paginator
	treeCreator *tree.TreeCreator
}

func (i *QueryFilter) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.paginator.CreatePagination(query)
	i.filter.CreateFilter(query)
	i.sorter.CreateOrder(query)
	i.treeCreator.CreateTree(query)
}

type fqKey struct{}

func (i *SQLHelper) InjectQueryFilter(c *gin.Context) error {
	q, err := i.CreateQueryFilter(c)
	if err != nil {
		return err
	}
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), fqKey{}, q))
	return err
}

func (i *SQLHelper) ExtractQueryFilter(ctx context.Context) *QueryFilter {
	if i, ok := ctx.Value(fqKey{}).(*QueryFilter); ok {
		return i
	}
	return nil
}
