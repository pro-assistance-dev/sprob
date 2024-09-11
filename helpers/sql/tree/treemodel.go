package tree

import (
	"encoding/json"
	"fmt"
	"strings"

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

func (i *TreeModel) CreateTree(query *bun.SelectQuery, relNames []string) {
	if i.Full {
		schema := project.SchemasLib.GetSchema(i.Model)
		fields := schema.GetFields()
		cols := make([]string, 0)
		relations := make([]project.Schema, 0)
		plural := false
		for _, field := range fields {
			var findedSchema project.Schema
			rel := project.SchemasLib.GetSchema(field)
			pluralRel := project.SchemasLib.GetSchemaByPluralName(field)
			if rel != nil {
				findedSchema = rel
			}
			if pluralRel != nil {
				plural = true
				findedSchema = pluralRel
			}
			if rel == nil {
				cols = append(cols, field)
			} else {
				relations = append(relations, findedSchema)
			}
		}

		colsForQuery := make([]string, 0)
		tableName := schema.GetTableName()
		for _, col := range cols {
			colName := schema.GetCol(col)
			if colName == "" {
				continue
			}
			colsForQuery = append(colsForQuery, fmt.Sprintf("%s.%s", tableName, colName))
		}
		for _, relation := range relations {
			relName := relation.GetScructName()
			if plural {
				relName = relation.GetPluralName()
			}
			relPath := append(relNames, relName)
			query.Relation(strings.Join(relPath, "."))
		}
		query.Column(colsForQuery...)

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
		node.CreateTree(query, nil)
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
