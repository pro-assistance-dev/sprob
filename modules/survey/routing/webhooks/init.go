package webhooks

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/webhooks"

	"github.com/gin-gonic/gin"
)

// Init — редактор вебхуков публикации (под /api/publications).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/:id/webhooks", h.GetAll)
	r.PUT("/:id/webhooks", h.ReplaceAll)
}
