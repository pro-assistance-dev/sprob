package public

import (
	handler "github.com/pro-assistance-dev/sprob/modules/survey/handlers/public"

	"github.com/gin-gonic/gin"
)

// Init — публичные ручки опросника (группа apiNoToken: без JWT).
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/:slug", h.GetForm)
	r.POST("/:slug/responses", h.Submit)
}
