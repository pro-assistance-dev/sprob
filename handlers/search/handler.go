package search

import (
	"encoding/json"
	"net/http"

	"github.com/pro-assistance-dev/sprob/models"

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

func (h *Handler) SearchMain(c *gin.Context) {
	var item models.SearchModel
	err := json.Unmarshal([]byte(c.Query("searchModel")), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = S.SearchMain(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
