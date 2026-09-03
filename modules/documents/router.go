package documents

import (
	"github.com/pro-assistance-dev/sprob/modules/documents/handlers/inns"
	"github.com/pro-assistance-dev/sprob/modules/documents/handlers/passports"
	"github.com/pro-assistance-dev/sprob/modules/documents/handlers/passportscans"
	"github.com/pro-assistance-dev/sprob/modules/documents/handlers/snilss"

	innsRouter "github.com/pro-assistance-dev/sprob/modules/documents/routing/inns"
	passportsRouter "github.com/pro-assistance-dev/sprob/modules/documents/routing/passports"
	passportscansRouter "github.com/pro-assistance-dev/sprob/modules/documents/routing/passportscans"
	snilssRouter "github.com/pro-assistance-dev/sprob/modules/documents/routing/snilss"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	passportsRouter.Init(api.Group("/passports"), passports.Init(helper))

	passportscansRouter.Init(api.Group("/passports-scans"), passportscans.Init(helper))

	snilssRouter.Init(api.Group("/snilss"), snilss.Init(helper))

	innsRouter.Init(api.Group("/inns"), inns.Init(helper))
}
