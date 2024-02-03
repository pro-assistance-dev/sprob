package t

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
)

// treeModel model
type treeModel struct {
	Model string   `json:"model"`
	Cols  []string `json:"col"`
}

type TreeModels []*treeModel

func (s *treeModel) getTableAndCols() string {
	schema := projecthelper.SchemasLib.GetSchema(s.Model)
	var result string
	for _, value := range s.Cols {
		result += fmt.Sprintf("%s ", value)
	}
	return fmt.Sprintf("%s %s", schema.GetTableName(), result)
}
