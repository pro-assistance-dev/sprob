package search

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Searc(c *gin.Context) {
	var item models.SearchModel
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = S.Search(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
