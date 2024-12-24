package buildings

import (
	"github.com/pro-assistance-dev/sprob/modules/buildings/handlers/buildings"
	"github.com/pro-assistance-dev/sprob/modules/buildings/handlers/entrances"
	"github.com/pro-assistance-dev/sprob/modules/buildings/handlers/floors"

	buildingsRouter "github.com/pro-assistance-dev/sprob/modules/buildings/routing/buildings"
	entrancesRouter "github.com/pro-assistance-dev/sprob/modules/buildings/routing/entrances"
	floorsRouter "github.com/pro-assistance-dev/sprob/modules/buildings/routing/floors"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	buildings.Init(helper)
	buildingsRouter.Init(api.Group("/buildings"), buildings.H)

	entrances.Init(helper)
	entrancesRouter.Init(api.Group("/entrances"), entrances.H)

	floors.Init(helper)
	floorsRouter.Init(api.Group("/floors"), floors.H)
}
