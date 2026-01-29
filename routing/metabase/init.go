package menus

import (
	handler "github.com/pro-assistance-dev/sprob/handlers/menus"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("/xlsx/:cardId", h.GetAll)
	r.GET("/xlsx/:cardId", h.GetAll)
}
