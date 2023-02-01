package sorter

import (
	"encoding/json"
	"fmt"

	"github.com/pro-assistance/pro-assister/projecthelper"
)

// sortModel model
type sortModel struct {
	Model   string `json:"model"`
	Table   string `json:"table"`
	Col     string `json:"col"`
	Order   Orders `json:"order"`
	Version string `json:"version"`
}

type SortModels []*sortModel

// parseJSONToSortModel constructor
func parseJSONToSortModel(args string) (sortModel sortModel, err error) {
	err = json.Unmarshal([]byte(args), &sortModel)
	if err != nil {
		return sortModel, err
	}
	return sortModel, err
}

func (s *sortModel) getTableAndCol() string {
	if s.Version == "v2" {
		schema := projecthelper.SchemasLib.GetSchema(s.Model)
		return fmt.Sprintf("%s.%s", schema.GetTableName(), schema.GetCol(s.Col))
	}
	return fmt.Sprintf("%s.%s", s.Table, s.Col)
}
