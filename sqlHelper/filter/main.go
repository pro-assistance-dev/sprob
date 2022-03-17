package filter

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type Filter struct {
	ID           *string
	FilterModels FilterModels
}

func NewFilter(c *gin.Context) (*Filter, error) {
	filterModels, err := createFilterModels(c)
	if err != nil {
		return nil, err
	}
	return &Filter{FilterModels: filterModels}, err
}

func createFilterModels(c *gin.Context) (FilterModels, error) {
	filterModels := make(FilterModels, 0)
	if c.Query("filterModel") == "" {
		return nil, nil
	}
	for _, arg := range c.QueryArray("filterModel") {
		filterModel, err := parseJSONToFilterModel(arg)
		if err != nil {
			return nil, err
		}
		filterModels = append(filterModels, &filterModel)
	}

	return filterModels, nil
}

// ParseJSONToFilterModel constructor
func parseJSONToFilterModel(args string) (filterModel FilterModel, err error) {
	err = json.Unmarshal([]byte(args), &filterModel)
	if err != nil {
		return filterModel, err
	}
	return filterModel, err
}
