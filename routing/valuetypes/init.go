package auth

import (
	"github.com/gin-gonic/gin"
	handler "github.com/pro-assistance-dev/sprob/handlers/valuetypes"
)

// Init func
func Init(r *gin.RouterGroup, h handler.IHandler) {
	r.GET("", h.GetAll)
	r.GET("/:id", h.Get)
}
