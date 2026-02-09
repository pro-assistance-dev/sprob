package metabase

import (
	handler "github.com/pro-assistance-dev/sprob/handlers/metabase"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/xlsx/:name", h.XLSX)
	r.GET("/frame/:questionId", h.Frame)
}
