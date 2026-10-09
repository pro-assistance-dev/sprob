package rules

import (
	rulesH "github.com/pro-assistance-dev/sprob/notify/handlers/rules"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.RouterGroup, h *rulesH.Handler) {
	r.GET("/options/:label/:value", h.Options)
	r.GET("", h.GetAll)
	r.POST("/ftsp", h.FTSP)
	r.GET("/:id", h.Get)
	r.POST("", h.Create)
	r.PUT("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
}
