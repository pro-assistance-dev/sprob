package routing

import (
	"github.com/pro-assistance-dev/sprob/handlers/auth"
	"github.com/pro-assistance-dev/sprob/handlers/fileinfos"
	"github.com/pro-assistance-dev/sprob/handlers/ftsppresets"
	"github.com/pro-assistance-dev/sprob/handlers/menus"
	"github.com/pro-assistance-dev/sprob/handlers/schemas"

	// "github.com/pro-assistance-dev/sprob/handlers/search"
	"github.com/pro-assistance-dev/sprob/handlers/usersaccounts"
	"github.com/pro-assistance-dev/sprob/handlers/valuetypes"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"

	"github.com/pro-assistance-dev/sprob/modules/buildings"
	"github.com/pro-assistance-dev/sprob/modules/chats"
	"github.com/pro-assistance-dev/sprob/modules/documents"
	"github.com/pro-assistance-dev/sprob/modules/extracts"
	"github.com/pro-assistance-dev/sprob/modules/forms"
	"github.com/pro-assistance-dev/sprob/modules/settings"

	"github.com/pro-assistance-dev/sprob/handlers/humans"
	fileinfosRouter "github.com/pro-assistance-dev/sprob/routing/fileinfos"
	ftsppresetsRouter "github.com/pro-assistance-dev/sprob/routing/ftsppresets"
	humansR "github.com/pro-assistance-dev/sprob/routing/humans"
	menusRouter "github.com/pro-assistance-dev/sprob/routing/menus"
	schemasRouter "github.com/pro-assistance-dev/sprob/routing/schemas"
	useraccountsRouter "github.com/pro-assistance-dev/sprob/routing/usersaccounts"
	valuetypesRouter "github.com/pro-assistance-dev/sprob/routing/valuetypes"

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

	// phones.Init(h)
	// phonesRouter.Init(apiToken.Group("/phones"), phones.H)

	humans.Init(h)
	humansR.Init(apiToken.Group("/humans"), humans.H)

	ftsppresets.Init(h)
	ftsppresetsRouter.Init(apiToken.Group("/ftsp-presets"), ftsppresets.H)

	schemas.Init(h)
	schemasRouter.Init(apiToken.Group("/schemas"), schemas.H)

	// emails.Init(h)
	// emailsRouter.Init(apiToken.Group("/emails"), emails.H)

	// contacts.Init(h)
	// contactsRouter.Init(apiToken.Group("/contacts"), contacts.H)

	menus.Init(h)
	menusRouter.Init(apiToken.Group("/menus"), menus.H)

	// metabase.Init(h)
	// metabaseR.Init(apiToken.Group("/metabase"), metabase.H)
	// search.Init(h)
	// searchRouter.Init(apiToken.Group("/search"), search.H)

	fileinfos.Init(h)
	fileinfosRouter.Init(apiToken.Group("/file-infos"), fileinfos.H)

	valuetypes.Init(h)
	valuetypesRouter.Init(apiToken.Group("/value-types"), valuetypes.H)

	usersaccounts.Init(h)
	useraccountsRouter.Init(apiToken.Group("/users-accounts"), usersaccounts.H)

	forms.InitRoutes(apiToken, h)
	settings.InitRoutes(apiToken, h)
	extracts.InitRoutes(apiToken, h)
	buildings.InitRoutes(apiToken, h)
	chats.InitRoutes(apiToken, h)
	documents.InitRoutes(apiToken, h)

	return apiToken, apiNoToken
}
