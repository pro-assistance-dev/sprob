package sqlHelper

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
)

type SQLHelper struct {
}

type fqKey struct{}

func NewSQLHelper() *SQLHelper {
	return &SQLHelper{}
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
	return &QueryFilter{Col: col, Value: value, filter: filterItem, sorter: sorterItem, paginator: paginatorItem}, nil
}

func (i *SQLHelper) WhereLikeWithLowerTranslit(col string, search string) string {
	return fmt.Sprintf("WHERE lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g')) LIKE lower(%s)", col, "'%"+search+"%'")
}

func (i *SQLHelper) InjectQueryFilter(c *gin.Context) (context.Context, error) {
	q, err := i.CreateQueryFilter(c)
	if err != nil {
		return nil, err
	}
	ctx := context.WithValue(c, fqKey{}, q)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), fqKey{}, q))
	return ctx, err
}

func (i *SQLHelper) ExtractQueryFilter(ctx context.Context) *QueryFilter {
	if i, ok := ctx.Value(fqKey{}).(*QueryFilter); ok {
		return i
	}
	return nil
}
