package models

import (
	"github.com/pro-assistance-dev/sprob/helpers/sql"
)

type FTSPQuery struct {
	QID  string   `json:"qid"`
	FTSP sql.FTSP `json:"ftsp"`
}
type FTSPAnswer struct {
	Data interface{} `json:"data"`
	FTSP sql.FTSP    `json:"ftsp"`
}
