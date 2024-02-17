package routing

import (
	"pro-assister/handlers/auth"
	"pro-assister/handlers/search"
	"pro-assister/helper"

	searchRouter "pro-assister/routing/search"

	"github.com/gin-gonic/gin"
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
	// authRouter.Init(api.Group("/auth"), auth.H)

	search.Init(h)
	searchRouter.Init(api.Group("/search"), search.H)
	return api
}
