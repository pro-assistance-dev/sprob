package paginator

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

func NewPaginator(c *gin.Context) (pagination *Paginator, err error) {
	paginationQuery := c.Query("pagination")
	if paginationQuery == "" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(paginationQuery), &pagination)
	if err != nil {
		return pagination, err
	}
	return pagination, err
}
