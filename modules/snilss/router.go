package snilss

import (
	"github.com/pro-assistance-dev/sprob/modules/snilss/handlers/snilss"
	snilssRouter "github.com/pro-assistance-dev/sprob/modules/snilss/routing/snilss"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	snilss.Init(helper)
	snilssRouter.Init(api.Group("/snilss"), snilss.H)
}
