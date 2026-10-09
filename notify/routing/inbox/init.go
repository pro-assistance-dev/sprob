package inbox

import (
	inboxH "github.com/pro-assistance-dev/sprob/notify/handlers/inbox"

	"github.com/gin-gonic/gin"
)

// Init — роуты «входящих» in-app уведомлений.
func Init(r *gin.RouterGroup, h *inboxH.Handler) {
	r.GET("", h.GetAll)
	r.POST("/read-all", h.ReadAll)
	r.POST("/:id/read", h.Read)
}
