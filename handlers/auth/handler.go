package auth

import (
	"net/http"
	"pro-assister/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CheckUUID(c *gin.Context) {
	err := S.CheckUUID(c.Request.Context(), c.Param("id"), c.Param("uuid"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (h *Handler) RefreshPassword(c *gin.Context) {
	var item models.UserAccount
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = S.UpdatePassword(c.Request.Context(), item.ID.UUID.String(), item.Password)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, nil)
}
