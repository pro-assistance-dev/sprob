package sorter

import "github.com/gin-gonic/gin"

func NewSorter(c *gin.Context) (*Sorter, error) {
	models, err := createSortModels(c)
	if err != nil {
		return nil, err
	}
	return &Sorter{sortModels: models}, err
}

func createSortModels(c *gin.Context) (SortModels, error) {
	sortModels := make(SortModels, 0)
	if c.Query("sortModel") == "" {
		return nil, nil
	}
	for _, arg := range c.QueryArray("sortModel") {
		sortModel, err := parseJSONToSortModel(arg)
		if err != nil {
			return nil, err
		}
		sortModels = append(sortModels, &sortModel)
	}

	return sortModels, nil
}
