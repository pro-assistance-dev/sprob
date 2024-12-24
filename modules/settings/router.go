package settings

import (
	"github.com/pro-assistance-dev/sprob/modules/settings/handlers/colorthemes"

	colorthemesRouter "github.com/pro-assistance-dev/sprob/modules/settings/routing/colorthemes"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	colorthemes.Init(helper)
	colorthemesRouter.Init(api.Group("/color-themes"), colorthemes.H)
}
