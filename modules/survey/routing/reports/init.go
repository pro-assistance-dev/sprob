package reports

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/reports"

	"github.com/gin-gonic/gin"
)

// Init — аналитика/выгрузка по публикации (под /api/publications).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/:id/summary", h.Summary)
	r.GET("/:id/export", h.Export)
	r.GET("/:id/responses", h.Responses)
	r.GET("/by-form/:formId", h.ByForm)
}
