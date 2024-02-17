package search

import (
	handler "github.com/pro-assistance/pro-assister/handlers/search"

	"github.com/gin-gonic/gin"
)

// Init func
func Init(r *gin.RouterGroup, h handler.IHandler) {
	r.GET("", h.Search)
}
