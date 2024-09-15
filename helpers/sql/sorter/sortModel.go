package sorter

import (
	"encoding/json"

	"github.com/pro-assistance/pro-assister/helpers/project"
)

// sortModel model
type sortModel struct {
	Model string `json:"model"`
	Table string `json:"table"`
	Col   string `json:"col"`
	Order Orders `json:"order"`
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
	return project.SchemasLib.GetSchema(s.Model).ConcatTableCol(s.Col)
}
