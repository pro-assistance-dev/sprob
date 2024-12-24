package inns

import (
	"github.com/pro-assistance-dev/sprob/modules/inns/handlers/inns"
	innsRouter "github.com/pro-assistance-dev/sprob/modules/inns/routing/inns"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	inns.Init(helper)
	innsRouter.Init(api.Group("/inns"), inns.H)
}
