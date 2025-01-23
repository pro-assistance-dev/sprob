package chats

import (
	"github.com/pro-assistance-dev/sprob/modules/chats/handlers/chats"
	"github.com/pro-assistance-dev/sprob/modules/chats/handlers/chatsmessages"
	"github.com/pro-assistance-dev/sprob/modules/chats/handlers/chatsusers"

	chatsRouter "github.com/pro-assistance-dev/sprob/modules/chats/routing/chats"
	chatsmessagesRouter "github.com/pro-assistance-dev/sprob/modules/chats/routing/chatsmessages"
	chatsusersRouter "github.com/pro-assistance-dev/sprob/modules/chats/routing/chatsusers"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	chats.Init(helper)
	chatsRouter.Init(api.Group("/chats"), chats.H)

	chatsmessages.Init(helper)
	chatsmessagesRouter.Init(api.Group("/chat-messages"), chatsmessages.H)

	chatsusers.Init(helper)
	chatsusersRouter.Init(api.Group("/chats-users"), chatsusers.H)
}
