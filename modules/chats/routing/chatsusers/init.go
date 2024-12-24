package chatsusers

import (
	handler "github.com/pro-assistance-dev/sprob/modules/chats/handlers/chatsusers"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("", h.GetAll)
}
