package notifications

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/notifications"

	"github.com/gin-gonic/gin"
)

// Init — редактор уведомлений публикации (под /api/publications).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/:id/notifications", h.GetAll)
	r.PUT("/:id/notifications", h.ReplaceAll)
	r.GET("/:id/notification-logs", h.Logs)
}
