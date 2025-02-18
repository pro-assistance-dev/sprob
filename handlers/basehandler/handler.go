package basehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler[TSingle, TPlural, TPluralWithCount]) Create(c *gin.Context) {
	var item TSingle
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = h.S.Create(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
