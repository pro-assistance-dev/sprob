package tree

import (
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

// treeModel model
type TreeModel struct {
	Model  string      `json:"model"`
	Cols   []string    `json:"cols"`
	Models []TreeModel `json:"models"`
	Full   bool        `json:"full"`

	relationPath string
}

func (i *TreeModel) CreateTree(query *bun.SelectQuery) {
	schema := project.SchemasLib.GetSchema(i.Model)
	if schema == nil {
		return
	}
	fieldsSchemas := schema.GetFieldsWithSchema()

	for _, relation := range fieldsSchemas {
		newRelation := relation.NamePascal
		if newRelation == "Children" {
			continue
		}
		if i.relationPath != "" {
			newRelation = strings.Join([]string{i.relationPath, relation.NamePascal}, ".")
		}

		// Глубина вложенности связей не ограничена — на циклических моделях
		// это зациклит запрос. `Full` задуман как рычаг, но пока не реализован.
		query.Relation(newRelation)

		typeString := strcase.ToLowerCamel(pluralize.NewClient().Singular(relation.NameCamel))

		treeModel := TreeModel{Model: typeString, relationPath: newRelation}
		treeModel.CreateTree(query)
	}
}
