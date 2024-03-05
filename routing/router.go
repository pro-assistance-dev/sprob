package routing

import (
	"github.com/pro-assistance/pro-assister/handlers/auth"
	"github.com/pro-assistance/pro-assister/handlers/fileinfos"
	"github.com/pro-assistance/pro-assister/handlers/search"
	"github.com/pro-assistance/pro-assister/handlers/valuetypes"
	"github.com/pro-assistance/pro-assister/helper"
	"github.com/pro-assistance/pro-assister/middleware"

	fileinfosRouter "github.com/pro-assistance/pro-assister/routing/fileinfos"
	searchRouter "github.com/pro-assistance/pro-assister/routing/search"
	valuetypesRouter "github.com/pro-assistance/pro-assister/routing/valuetypes"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, h *helper.Helper) (*gin.RouterGroup, *gin.RouterGroup) {
	m := middleware.CreateMiddleware(h)
	r.Use(m.CORSMiddleware())
	r.Use(gin.Logger())

	r.Static("/api/static", "./static/")

	apiToken := r.Group("/api")
	apiToken.Use(m.InjectRequestInfo())

	apiNoToken := r.Group("/api")

	auth.Init(h)
	// authRouter.Init(api.Group("/auth"), auth.H)

	search.Init(h)
	searchRouter.Init(apiToken.Group("/search"), search.H)

	fileinfos.Init(h)
	fileinfosRouter.Init(apiToken.Group("/file-infos"), fileinfos.H)

	valuetypes.Init(h)
	valuetypesRouter.Init(apiToken.Group("/value-types"), valuetypes.H)
	return apiToken, apiNoToken
}
