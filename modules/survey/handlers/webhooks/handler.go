package webhooks

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
)

// Handler — редактор вебхуков публикации.
type Handler struct {
	helper *helper.Helper
	s      *Service
}

var H *Handler

// Init собирает цепочку handler→service→repository.
func Init(h *helper.Helper) *Handler {
	r := &Repository{db: h.DB.DB}
	s := &Service{r: r}
	handler := &Handler{helper: h, s: s}
	H = handler
	return handler
}

// GetAll — GET /api/publications/:id/webhooks
func (h *Handler) GetAll(c *gin.Context) {
	res, err := h.s.GetAll(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// ReplaceAll — PUT /api/publications/:id/webhooks — сохранить весь список.
func (h *Handler) ReplaceAll(c *gin.Context) {
	var payload struct {
		Items models.Webhooks `json:"items"`
	}
	_, err := h.helper.HTTP.GetForm(c, &payload)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	res, err := h.s.ReplaceAll(c.Request.Context(), c.Param("id"), payload.Items)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}
