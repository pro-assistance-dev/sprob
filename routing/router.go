package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/handlers/auth"
	"github.com/pro-assistance/pro-assister/helper"
	authRouter "github.com/pro-assistance/pro-assister/routing/auth"
)

func Init(r *gin.Engine, h *helper.Helper) *gin.RouterGroup {
	// m := middleware.CreateMiddleware(helper)
	// r.Use(m.InjectFTSP())
	// r.Use(m.CORSMiddleware())
	// r.Use(m.CheckPermission())
	r.Use(gin.Logger())
	r.Static("/api/static", "./static/")

	api := r.Group("/api")

	auth.Init(h)
	authRouter.Init(api.Group("/auth"), auth.H)

	return api
}
