package tree

import (
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
)

// treeModel model
type TreeModel struct {
	Model string   `json:"model"`
	Cols  []string `json:"col"`
}

type TreeModels []*TreeModel

func (s TreeModels) getTableAndCols() string {
	schema := projecthelper.SchemasLib.GetSchema(s[0].Model)
	// var result string
	// for _, value := range s.Cols {
	// 	result += fmt.Sprintf("%s ", value)
	// }
	// fmt.Println(schema.GetTableName(), result)
	return fmt.Sprintf("%s.%s", schema.GetTableName(), schema.GetCol(s[0].Cols[0]))
}
