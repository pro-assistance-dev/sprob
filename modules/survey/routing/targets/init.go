package targets

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/targets"

	"github.com/gin-gonic/gin"
)

// Init — привязка публикаций к сущностям экосистемы (под /api/publications).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/by-target/:type/:id", h.List)
	r.POST("/for-target", h.Ensure)
}
