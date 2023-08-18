package sqlHelper

import (
	"context"
	"fmt"
)

type SQLHelper struct {
}

type fqKey struct{}

func NewSQLHelper() *SQLHelper {
	return &SQLHelper{}
}

func (i *SQLHelper) WhereLikeWithLowerTranslit(col string, search string) string {
	return fmt.Sprintf("WHERE lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g')) LIKE lower(%s)", col, "'%"+search+"%'")
}

func (i *SQLHelper) InjectQueryFilter(c context.Context, q *QueryFilter) {
	c = context.WithValue(c, fqKey{}, q)
}

func ExtractQueryFilter(ctx context.Context) *QueryFilter {
	if i, ok := ctx.Value(fqKey{}).(*QueryFilter); ok {
		return i
	}
	return nil
}
