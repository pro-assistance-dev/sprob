package tree

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pro-assistance/pro-assister/helpers/project"
)

// treeModel model
type TreeModel struct {
	Model  string      `json:"model"`
	Cols   []string    `json:"col"`
	Models []TreeModel `json:"models"`
}

func parseJSONToTreeModel(args string) (treeModel TreeModel, err error) {
	err = json.Unmarshal([]byte(args), &treeModel)
	if err != nil {
		return treeModel, err
	}
	return treeModel, err
}

func (t *TreeModel) getTableAndCols() string {
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
	return strings.Join(cols, ", ")
}
