package paginator

import (
	"github.com/uptrace/bun"
)

type Paginator struct {
	Page        *int    `json:"page"`
	RowsPerPage *int    `json:"rowsPerPage"`
	CursorMode  bool    `json:"cursorMode"`
	Cursor      *Cursor `json:"cursor"`
	TableName   string  `json:"tableName"`
}

func (i *Paginator) CreatePagination(query *bun.SelectQuery) {
	if i == nil {
		return
	}
	if i.CursorMode {
		i.Cursor.createPagination(query)
	} else {
		query = query.Offset(*i.Page)
	}
	query.Limit(*i.RowsPerPage)
}
