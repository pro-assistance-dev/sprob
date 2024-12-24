package passports

import (
	"github.com/pro-assistance-dev/sprob/modules/passports/handlers/passports"
	"github.com/pro-assistance-dev/sprob/modules/passports/handlers/passportscans"

	passportsRouter "github.com/pro-assistance-dev/sprob/modules/passports/routing/passports"
	passportscansRouter "github.com/pro-assistance-dev/sprob/modules/passports/routing/passportscans"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	passports.Init(helper)
	passportsRouter.Init(api.Group("/passports"), passports.H)

	passportscans.Init(helper)
	passportscansRouter.Init(api.Group("/passports-scans"), passportscans.H)
}
