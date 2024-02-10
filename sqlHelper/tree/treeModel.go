package tree

import (
	"encoding/json"
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
)

// treeModel model
type treeModel struct {
	Model string   `json:"model"`
	Cols  []string `json:"col"`
}

type TreeModels []*treeModel

func parseJSONToTreeModel(args string) (treeModel treeModel, err error) {
	err = json.Unmarshal([]byte(args), &treeModel)
	if err != nil {
		return treeModel, err
	}
	return treeModel, err
}

func (s *treeModel) getTableAndCols() string {
	schema := projecthelper.SchemasLib.GetSchema(s.Model)
	return fmt.Sprintf("%s.%s", schema.GetTableName(), schema.GetCol(s.Cols[0]))
}
