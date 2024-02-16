package sql

import (
	"fmt"

	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type SQL struct{}

func NewSQL() *SQL {
	return &SQL{}
}

func (i *SQL) WhereLikeWithLowerTranslit(col string, search string) string {
	return fmt.Sprintf("WHERE lower(regexp_replace(%s, '[^а-яА-Яa-zA-Z0-9 ]', '', 'g')) LIKE lower(%s)", col, "'%"+search+"%'")
}

func (i *SQL) HandleFTSPQuery(ctx context.Context, query *bun.SelectQuery) {
	i.ExtractFTSP(ctx).HandleQuery(query)
}
