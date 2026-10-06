package invites

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/invites"

	"github.com/gin-gonic/gin"
)

// Init — рассылка приглашений по публикации (под /api/publications).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.POST("/:id/invites/send", h.Send)
}
