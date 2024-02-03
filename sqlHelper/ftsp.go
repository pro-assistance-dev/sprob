package sqlHelper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/sqlHelper/filter"
	"github.com/pro-assistance/pro-assister/sqlHelper/paginator"
	"github.com/pro-assistance/pro-assister/sqlHelper/sorter"
	"github.com/pro-assistance/pro-assister/sqlHelper/tree"
	"github.com/uptrace/bun"
)

type FTSP struct {
	ID    string               `json:"id"`
	Col   string               `json:"col"`
	Value string               `json:"value"`
	F     filter.FilterModels  `json:"f"`
	T     tree.TreeModel       `json:"t"`
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
	i.T.CreateTree(query)
}

type ftspKey struct{}

type FTSPQuery struct {
	QID  string `json:"qid"`
	FTSP FTSP   `json:"ftsp"`
}

func (i *SQLHelper) InjectFTSP2(r *http.Request, f *FTSP) {
	*r = *r.WithContext(context.WithValue(r.Context(), ftspKey{}, f))
	fmt.Println(r)
}

func (i *SQLHelper) InjectFTSP(c *gin.Context) error {
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

func (i *SQLHelper) ExtractFTSP(ctx context.Context) *FTSP {
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
