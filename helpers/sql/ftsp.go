package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pro-assistance/pro-assister/helpers/project"
	"github.com/pro-assistance/pro-assister/helpers/sql/filter"
	"github.com/pro-assistance/pro-assister/helpers/sql/paginator"
	"github.com/pro-assistance/pro-assister/helpers/sql/sorter"
	"github.com/pro-assistance/pro-assister/helpers/sql/tree"
	"github.com/uptrace/bun"
)

type FTSP struct {
	bun.BaseModel `bun:"ftsp,alias:ftsp"`
	ID            uuid.NullUUID        `json:"id"`
	Col           string               `json:"col"`
	Value         string               `json:"value"`
	F             filter.FilterModels  `json:"f"`
	T             tree.TreeModels      `json:"t"` // добавил
	S             sorter.SortModels    `json:"s"`
	P             *paginator.Paginator `json:"p"`
}

func (i *FTSP) HandleQuery(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	i.distinctOn(query)
	i.P.CreatePagination(query)
	i.F.CreateFilter(query)
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
		sortCol := t.GetCol(i.S[0].Col)
		query.DistinctOn(fmt.Sprintf("%s.%s, %s.id", t.GetTableName(), sortCol, t.GetTableName()))
	}
}
func (i *SQL) InjectFTSP2(r *http.Request, f *FTSP) {
	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, f))
	fmt.Println(r)
}

func (i *SQL) InjectFTSP(c *gin.Context) error {
	ftsp := &FTSPQuery{}
	err := ftsp.FromForm(c)
	fmt.Println("ftsp", ftsp)
	if err != nil {
		fmt.Println(err)
		return err
	}
	r := c.Request

	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, ftsp.FTSP))
	// c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ftspKey{}, ftsp.FTSP))
	return err
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
	err = json.Unmarshal([]byte(form.Value["form"][0]), i)
	if err != nil {
		return err
	}
	return nil
}
