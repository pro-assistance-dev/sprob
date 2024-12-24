package formsections

import (
	handler "github.com/pro-assistance-dev/sprob/modules/forms/handlers/formsections"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("", h.GetAll)
	r.GET("/:id", h.Get)
	r.POST("", h.Create)
	r.DELETE("/:id", h.Delete)
	r.PUT("/many", h.UpdateMany)
	r.PUT("/:id", h.Update)
}
