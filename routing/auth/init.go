package auth

import (
	"github.com/gin-gonic/gin"
	handler "github.com/pro-assistance-dev/sprob/handlers/auth"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.PUT("/refresh-password", h.RefreshPassword)
	r.GET("/confirm-email/:id", h.ConfirmEmail)
	r.GET("/email-is-confirm/:email", h.EmailIsConfirm)
	r.GET("/check-uuid/:id/:uuid", h.CheckUUID)
}
