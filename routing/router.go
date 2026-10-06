package routing

import (
	"github.com/pro-assistance-dev/sprob/handlers/auth"
	"github.com/pro-assistance-dev/sprob/handlers/fileinfos"
	"github.com/pro-assistance-dev/sprob/handlers/menus"
	"github.com/pro-assistance-dev/sprob/handlers/schemas"

	// "github.com/pro-assistance-dev/sprob/handlers/search"
	"github.com/pro-assistance-dev/sprob/handlers/usersaccounts"
	"github.com/pro-assistance-dev/sprob/handlers/valuetypes"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/middleware"

	"github.com/pro-assistance-dev/sprob/handlers/metabase"
	"github.com/pro-assistance-dev/sprob/modules/chats"
	"github.com/pro-assistance-dev/sprob/modules/documents"
	"github.com/pro-assistance-dev/sprob/modules/extracts"
	"github.com/pro-assistance-dev/sprob/modules/forms"
	"github.com/pro-assistance-dev/sprob/modules/settings"
	"github.com/pro-assistance-dev/sprob/modules/survey"
	surveyModels "github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/pro-assistance-dev/sprob/handlers/humans"
	fileinfosRouter "github.com/pro-assistance-dev/sprob/routing/fileinfos"
	humansR "github.com/pro-assistance-dev/sprob/routing/humans"
	menusRouter "github.com/pro-assistance-dev/sprob/routing/menus"
	metabaseR "github.com/pro-assistance-dev/sprob/routing/metabase"
	schemasRouter "github.com/pro-assistance-dev/sprob/routing/schemas"
	useraccountsRouter "github.com/pro-assistance-dev/sprob/routing/usersaccounts"
	valuetypesRouter "github.com/pro-assistance-dev/sprob/routing/valuetypes"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, h *helper.Helper) (*gin.RouterGroup, *gin.RouterGroup) {
	m := middleware.CreateMiddleware(h)
	r.Use(m.CORSMiddleware())
	r.Use(gin.Logger())
	uploaderPath := h.Uploader.GetUploaderPath()
	if uploaderPath != nil {
		r.Static("/api/static", *uploaderPath)
	}

	apiToken := r.Group("/api")
	apiToken.Use(m.InjectRequestInfo())

	apiNoToken := r.Group("/api")

	auth.Init(h)
	// authRouter.Init(api.Group("/auth"), auth.H)

	// phones.Init(h)
	// phonesRouter.Init(apiToken.Group("/phones"), phones.H)

	humansR.Init(apiToken.Group("/humans"), humans.Init(h))

	schemasRouter.Init(apiToken.Group("/schemas"), schemas.Init(h))

	// emails.Init(h)
	// emailsRouter.Init(apiToken.Group("/emails"), emails.H)

	// contacts.Init(h)
	// contactsRouter.Init(apiToken.Group("/contacts"), contacts.H)

	menusRouter.Init(apiToken.Group("/menus"), menus.Init(h))

	metabaseR.Init(apiToken.Group("/metabase"), metabase.Init(h))
	// search.Init(h)
	// searchRouter.Init(apiToken.Group("/search"), search.H)

	fileinfosRouter.Init(apiToken.Group("/file-infos"), fileinfos.Init(h))

	valuetypesRouter.Init(apiToken.Group("/value-types"), valuetypes.Init(h))

	useraccountsRouter.Init(apiToken.Group("/users-accounts"), usersaccounts.Init(h))

	forms.InitRoutes(apiToken, h)
	settings.InitRoutes(apiToken, h)
	extracts.InitRoutes(apiToken, h)
	chats.InitRoutes(apiToken, h)
	documents.InitRoutes(apiToken, h)

	// Модуль survey (опросник): CRUD+FTSP доменных моделей здесь (импорт
	// sprob/routing из модуля дал бы цикл), публичный рантайм и свои ручки — внутри.
	survey.InitRoutes(apiToken, apiNoToken, h)
	InitR[surveyModels.Publication](apiToken)
	InitR[surveyModels.SurveyParam](apiToken)
	InitR[surveyModels.Response](apiToken)
	InitR[surveyModels.Invite](apiToken)
	InitR[surveyModels.Notification](apiToken)
	InitR[surveyModels.NotificationRule](apiToken)
	InitR[surveyModels.NotificationLog](apiToken)
	InitR[surveyModels.Theme](apiToken)
	InitR[surveyModels.Webhook](apiToken)
	InitR[surveyModels.WebhookRule](apiToken)
	InitR[surveyModels.WebhookLog](apiToken)

	return apiToken, apiNoToken
}
