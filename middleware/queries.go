package middleware

import (
	"context"
	"net/http"

	"github.com/pro-assistance/pro-assister/helpers/sql"
)

type Query struct {
	QID  string    `json:"qid"`
	FTSP *sql.FTSP `json:"ftsp"`
}

var queriesMap = make(map[string]*sql.FTSP)

const ftsp = "ftsp"

func (item Query) Inject(r *http.Request, qid string) error {
	ftsp := getFTSP(qid)
	// if ftsp == nil {
	// 	setFTSP(nil)
	// }
	*r = *r.WithContext(context.WithValue(r.Context(), ftsp, ftsp))
	return nil
}

func getFTSP(qid string) *sql.FTSP {
	filter, ok := queriesMap[qid]
	if !ok {
		return nil
	}
	return filter
}

func setFTSP(query Query) {
	queriesMap[query.QID] = query.FTSP
}
