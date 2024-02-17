package fileinfos

import (
	handler "pro-assister/handlers/fileinfos"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h handler.IHandler) {
	r.GET("/:id", h.Download)
	r.POST("", h.Create)
}
