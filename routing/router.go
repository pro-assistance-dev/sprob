package routing

import (
	"github.com/pro-assistance/pro-assister/handlers/auth"
	"github.com/pro-assistance/pro-assister/handlers/contacts"
	"github.com/pro-assistance/pro-assister/handlers/emails"
	"github.com/pro-assistance/pro-assister/handlers/fileinfos"
	"github.com/pro-assistance/pro-assister/handlers/ftsppresets"
	"github.com/pro-assistance/pro-assister/handlers/menus"
	"github.com/pro-assistance/pro-assister/handlers/phones"
	"github.com/pro-assistance/pro-assister/handlers/search"
	"github.com/pro-assistance/pro-assister/handlers/usersaccounts"
	"github.com/pro-assistance/pro-assister/handlers/valuetypes"
	"github.com/pro-assistance/pro-assister/helper"
	"github.com/pro-assistance/pro-assister/middleware"

	contactsRouter "github.com/pro-assistance/pro-assister/routing/contacts"
	emailsRouter "github.com/pro-assistance/pro-assister/routing/emails"
	fileinfosRouter "github.com/pro-assistance/pro-assister/routing/fileinfos"
	ftsppresetsRouter "github.com/pro-assistance/pro-assister/routing/ftsppresets"
	menusRouter "github.com/pro-assistance/pro-assister/routing/menus"
	phonesRouter "github.com/pro-assistance/pro-assister/routing/phones"
	searchRouter "github.com/pro-assistance/pro-assister/routing/search"
	useraccountsRouter "github.com/pro-assistance/pro-assister/routing/usersaccounts"
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

	phones.Init(h)
	phonesRouter.Init(apiToken.Group("/phones"), phones.H)

	ftsppresets.Init(h)
	ftsppresetsRouter.Init(apiToken.Group("/ftsp-presets"), ftsppresets.H)

	emails.Init(h)
	emailsRouter.Init(apiToken.Group("/emails"), emails.H)

	contacts.Init(h)
	contactsRouter.Init(apiToken.Group("/contacts"), contacts.H)

	menus.Init(h)
	menusRouter.Init(apiToken.Group("/menus"), menus.H)

	search.Init(h)
	searchRouter.Init(apiToken.Group("/search"), search.H)

	fileinfos.Init(h)
	fileinfosRouter.Init(apiToken.Group("/file-infos"), fileinfos.H)

	valuetypes.Init(h)
	valuetypesRouter.Init(apiToken.Group("/value-types"), valuetypes.H)

	usersaccounts.Init(h)
	useraccountsRouter.Init(apiToken.Group("/users-accounts"), usersaccounts.H)

	return apiToken, apiNoToken
}
