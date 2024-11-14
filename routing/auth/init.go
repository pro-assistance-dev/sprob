package auth

import (
	"github.com/gin-gonic/gin"
	handler "github.com/pro-assistance-dev/sprob/handlers/auth"
)

// Init func
func Init(r *gin.RouterGroup, h handler.IHandler) {
	r.PUT("/refresh-password", h.RefreshPassword)
	r.GET("/check-uuid/:id/:uuid", h.CheckUUID)
}
