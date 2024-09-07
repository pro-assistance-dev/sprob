package tree

import (
	"encoding/json"
	"fmt"

	"github.com/pro-assistance/pro-assister/helpers/project"
	"github.com/uptrace/bun"
)

// treeModel model
type TreeModel struct {
	Model  string      `json:"model"`
	Cols   []string    `json:"col"`
	Models []TreeModel `json:"models"`
	Full   bool        `json:"full"`
}

func parseJSONToTreeModel(args string) (treeModel TreeModel, err error) {
	err = json.Unmarshal([]byte(args), &treeModel)
	if err != nil {
		return treeModel, err
	}
	return treeModel, err
}

func (t *TreeModel) getTableAndCols() []string {
	schema := project.SchemasLib.GetSchema(t.Model)
	tableName := schema.GetTableName()
	cols := make([]string, 0)
	for i := range t.Cols {
		colName := schema.GetCol(t.Cols[i])
		if colName == "" {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s.%s", tableName, colName))
	}
	return cols
}

func (i *TreeModel) CreateTree(query *bun.SelectQuery) {
	if i.Full {
		// schema := project.SchemasLib.GetSchema(i.Model)
		// fields := schema.GetFields()
		// cols :=
		// for _, field := range fields {
		//
		// }
		return
	}
	// if len(i.Cols) == 0 && len(i.Models) == 0 {
	// 	return
	// }
	cols := i.getTableAndCols()
	if len(cols) > 0 {
		query.Column(cols...)
	}

	for _, node := range i.Models {
		node.CreateTree(query)
	}
}

//
// func (r *Repository) Get(c context.Context, id string) (*models.Form, error) {
// 	item := models.Form{}
// 	err := r.helper.DB.IDB(c).NewSelect().
// 		Model(&item).
// 		Relation("FormSections", func(q *bun.SelectQuery) *bun.SelectQuery {
// 			return q.Order("form_sections.item_order")
// 		}).
// 		Relation("FormSections.Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
// 			return q.Order("fields.item_order")
// 		}).
// 		Relation("FormSections.Fields.ValueType").
// 		Relation("FormSections.Fields.AnswerVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
// 			return q.Order("answer_variants.item_order")
// 		}).
// 		Where("?TableAlias.id = ?", id).Scan(c)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &item, err
// }
