package search

import (
	"net/http"

	"github.com/pro-assistance/pro-assister/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Search(c *gin.Context) {
	var item models.SearchModel
	item.Key = c.Query("key")
	item.Query = c.Query("query")
	err := S.Search(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
