package f

import (
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

// Model model
type Model struct {
	Model  string `json:"model"`
	Wheres Wheres `json:"wheres"`
	Joins  Joins  `json:"joins"`
}

// Models model
type Models []*Model

func (items Models) Filter(query *bun.SelectQuery) {
	if len(items) == 0 {
		return
	}
	for _, item := range items {
		item.Filter(query)
	}
}

func (item *Model) Filter(query *bun.SelectQuery) {
	schema := project.SchemasLib.GetSchema(item.Model)

	for _, where := range item.Wheres {
		where.Construct(schema.GetTableName(), *schema)
		where.Where(query)
	}
}
