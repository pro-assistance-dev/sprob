package invites

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
)

// Handler — рассылка приглашений по публикации.
type Handler struct {
	helper *helper.Helper
	s      *Service
}

var H *Handler

// Init собирает цепочку handler→service→repository.
func Init(h *helper.Helper) *Handler {
	handler := &Handler{helper: h, s: NewService(h.DB.DB)}
	H = handler
	return handler
}

// Send — POST /api/publications/:id/invites/send — создать приглашения и выслать письма.
// Тело: { recipients: [{email, fio}], subject?, body? }.
func (h *Handler) Send(c *gin.Context) {
	var payload struct {
		Recipients []Recipient `json:"recipients"`
		Subject    string      `json:"subject"`
		Body       string      `json:"body"`
	}
	_, err := h.helper.HTTP.GetForm(c, &payload)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	res, err := h.s.Send(c.Request.Context(), c.Param("id"), payload.Recipients, payload.Subject, payload.Body)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": res})
}
