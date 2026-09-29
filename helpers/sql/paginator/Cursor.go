package paginator

import (
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/sql/filter"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type Cursor struct {
	Operator  filter.Operator `json:"operation"`
	Column    string          `json:"column"`
	Value     string          `json:"value"`
	TableName string          `json:"tableName"`
	Model     string          `json:"model"`
	Initial   bool            `json:"initial"`
}

func (c *Cursor) createPagination(query *bun.SelectQuery) {
	if c.Initial {
		return
	}
	// ⚠️ Значение (`c.Value` — из клиента) — ПАРАМЕТР, а не `'%s'` в строке.
	// Колонка резолвится через схему модели → `bun.Ident`.
	if len(c.TableName) > 0 {
		query.Where("? ? ?", bun.Ident(c.getTableAndCol()), schema.SafeQuery(string(c.Operator), nil), c.Value)
		return
	}
	schemaModel := project.SchemasLib.GetSchema(c.Model)
	if schemaModel == nil {
		return
	}
	query.Where("? ? ?", bun.Ident(schemaModel.GetColName(c.Column)), schema.SafeQuery(string(c.Operator), nil), c.Value)
}

func (c *Cursor) getTableAndCol() string {
	return project.SchemasLib.GetSchema(c.Model).ConcatTableCol(c.Column)
}
