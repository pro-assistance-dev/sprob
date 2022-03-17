package sorter

import (
	"encoding/json"
	"fmt"
)

// sortModel model
type sortModel struct {
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
	return fmt.Sprintf("%s.%s", s.Table, s.Col)
}
