package tree

import (
	"encoding/json"
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

	relationPath string
}

func parseJSONToTreeModel(args string) (treeModel TreeModel, err error) {
	err = json.Unmarshal([]byte(args), &treeModel)
	if err != nil {
		return treeModel, err
	}
	return treeModel, err
}

func (i *TreeModel) CreateTree(query *bun.SelectQuery) {
	schema := project.SchemasLib.GetSchema(i.Model)

	fieldsCols := schema.ConcatTableCols()
	query.Column(fieldsCols...)

	fieldsSchemas := schema.GetFieldsWithSchema()

	for _, relation := range fieldsSchemas {
		newRelation := strings.Join([]string{i.relationPath, relation.NamePascal}, ".")
		query.Relation(newRelation)

		treeModel := TreeModel{Model: relation.NamePascal, relationPath: newRelation}
		treeModel.CreateTree(query)
	}
}
