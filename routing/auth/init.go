package auth

import (
	"github.com/gin-gonic/gin"
	handler "github.com/pro-assistance-dev/sprob/handlers/auth"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.PUT("/refresh-password", h.RefreshPassword)
	r.PUT("/confirm-email/:id", h.ConfirmEmail)
	r.GET("/check-uuid/:id/:uuid", h.CheckUUID)
}
