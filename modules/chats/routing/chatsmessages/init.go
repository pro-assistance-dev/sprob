package chatsmessages

import (
	handler "github.com/pro-assistance-dev/sprob/modules/chats/handlers/chatsmessages"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("", h.GetAll)
	r.GET("/:id", h.Get)
	r.POST("/ftsp", h.FTSP)
	r.POST("", h.Create)
	r.DELETE("/:id", h.Delete)
	r.PUT("/:id", h.Update)
}
