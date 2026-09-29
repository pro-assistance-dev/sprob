package sql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/sql/filter"
	"github.com/pro-assistance-dev/sprob/helpers/sql/filter/f"
	"github.com/pro-assistance-dev/sprob/helpers/sql/paginator"
	"github.com/pro-assistance-dev/sprob/helpers/sql/sorter"
	"github.com/pro-assistance-dev/sprob/helpers/sql/tree"
	"github.com/uptrace/bun"
)

type FTSP struct {
	Col   string               `json:"col"`
	Value string               `json:"value"`
	F     filter.FilterModels  `json:"f"`
	F2    f.Models             `json:"f2"`
	T     tree.TreeModel       `json:"t"`
	S     sorter.SortModels    `json:"s"`
	P     *paginator.Paginator `json:"p"`
}

func (i *FTSP) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.distinctOn(query)
	i.P.CreatePagination(query)
	i.F.CreateFilter(query)
	i.F2.Filter(query)
	i.S.CreateOrder(query)
	i.T.CreateTree(query)
}

type ftspKey struct{}

type FTSPQuery struct {
	QID  string `json:"qid"`
	FTSP FTSP   `json:"ftsp"`
}

func (i *FTSP) distinctOn(query *bun.SelectQuery) {
	if len(i.S) > 0 {
		t := project.SchemasLib.GetSchema(i.S[0].Model)
		sortCol := t.GetColName(i.S[0].Col)
		query.DistinctOn(fmt.Sprintf("%s.%s, %s.id", t.GetTableName(), sortCol, t.GetTableName()))
	}
}

func (i *SQL) InjectFTSP2(r *http.Request, f *FTSP) {
	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, f))
}

func (i *SQL) InjectFTSP(c *gin.Context) error {
	ftsp := &FTSPQuery{}
	if err := ftsp.FromForm(c); err != nil {
		return err
	}
	r := c.Request

	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, ftsp.FTSP))
	return nil
}

func (i *SQL) ExtractFTSP(ctx context.Context) *FTSP {
	if i, ok := ctx.Value(ftspKey{}).(*FTSP); ok {
		return i
	}
	return nil
}

func (i *FTSPQuery) FromForm(c *gin.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	// ⚠️ Без проверки `form.Value["form"][0]` паниковал на запросе без поля
	// `form` (индекс за границей среза) — клиенту уходил 500 вместо 400.
	values := form.Value["form"]
	if len(values) == 0 || values[0] == "" {
		return errors.New("пустое поле `form` в запросе")
	}
	if err := json.Unmarshal([]byte(values[0]), i); err != nil {
		return err
	}
	return nil
}
