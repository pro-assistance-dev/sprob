package routing

import (
	"mdgkb/prosodeystvie-server/handlers/auth"
	"mdgkb/prosodeystvie-server/handlers/banners"
	"mdgkb/prosodeystvie-server/handlers/contacts"
	"mdgkb/prosodeystvie-server/handlers/customsections"
	"mdgkb/prosodeystvie-server/handlers/emails"
	"mdgkb/prosodeystvie-server/handlers/event"
	"mdgkb/prosodeystvie-server/handlers/eventdays"
	"mdgkb/prosodeystvie-server/handlers/eventmessages"
	"mdgkb/prosodeystvie-server/handlers/experiences"
	"mdgkb/prosodeystvie-server/handlers/fileinfos"
	"mdgkb/prosodeystvie-server/handlers/humans"
	"mdgkb/prosodeystvie-server/handlers/m2m"
	"mdgkb/prosodeystvie-server/handlers/menus"
	"mdgkb/prosodeystvie-server/handlers/perfoms"
	"mdgkb/prosodeystvie-server/handlers/phones"
	"mdgkb/prosodeystvie-server/handlers/place"
	"mdgkb/prosodeystvie-server/handlers/schedules"
	"mdgkb/prosodeystvie-server/handlers/search"
	"mdgkb/prosodeystvie-server/handlers/sessions"
	"mdgkb/prosodeystvie-server/handlers/speakers"
	"mdgkb/prosodeystvie-server/handlers/sponsors"
	"mdgkb/prosodeystvie-server/handlers/users"
	"mdgkb/prosodeystvie-server/handlers/userseventssactivities"
	"mdgkb/prosodeystvie-server/middleware"
	authRouter "mdgkb/prosodeystvie-server/routing/auth"
	bannersRouter "mdgkb/prosodeystvie-server/routing/banners"
	contactsRouter "mdgkb/prosodeystvie-server/routing/contacts"
	customsectionsRouter "mdgkb/prosodeystvie-server/routing/customsections"
	emailsRouter "mdgkb/prosodeystvie-server/routing/emails"
	eventdaysRouter "mdgkb/prosodeystvie-server/routing/eventdays"
	eventMessagesRouter "mdgkb/prosodeystvie-server/routing/eventmessages"
	eventsRouter "mdgkb/prosodeystvie-server/routing/events"
	experiencesRouter "mdgkb/prosodeystvie-server/routing/experiences"
	fileinfosRouter "mdgkb/prosodeystvie-server/routing/fileinfos"
	humansRouter "mdgkb/prosodeystvie-server/routing/humans"
	m2mRouter "mdgkb/prosodeystvie-server/routing/m2m"
	menusRouter "mdgkb/prosodeystvie-server/routing/menus"
	perfomsRouter "mdgkb/prosodeystvie-server/routing/perfoms"
	phonesRouter "mdgkb/prosodeystvie-server/routing/phones"
	placeRouter "mdgkb/prosodeystvie-server/routing/place"
	schedulesRouter "mdgkb/prosodeystvie-server/routing/schedules"
	searchRouter "mdgkb/prosodeystvie-server/routing/search"
	sessionsRouter "mdgkb/prosodeystvie-server/routing/sessions"
	speakersRouter "mdgkb/prosodeystvie-server/routing/speakers"
	sponsorsRouter "mdgkb/prosodeystvie-server/routing/sponsors"
	usersRouter "mdgkb/prosodeystvie-server/routing/users"
	usersEventsActivitiesRouter "mdgkb/prosodeystvie-server/routing/userseventsactivities"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance/pro-assister/helper"
)

func Init(r *gin.Engine, helper *helperPack.Helper) {
	m := middleware.CreateMiddleware(helper)
	r.Use(m.InjectFTSP())
	// r.Use(m.CORSMiddleware())
	// r.Use(m.CheckPermission())
	r.Use(gin.Logger())
	// createdXlsxHelper := xlsxHelper.CreateXlsxHelper()
	r.Static("/api/static", "./static/")
	// r.Static("/static", "./static/")
	// r.Use(helper.HTTP.CORSMiddleware())
	api := r.Group("/api")
	authRouter.Init(api.Group("/auth"), auth.CreateHandler(helper))

	event.Init(helper)
	eventsRouter.Init(api.Group("/events"), event.H)

	sponsors.Init(helper)
	sponsorsRouter.Init(api.Group("/sponsors"), sponsors.H)

	banners.Init(helper)
	bannersRouter.Init(api.Group("/banners"), banners.H)

	fileinfosRouter.Init(api.Group("/file-infos"), fileinfos.CreateHandler(helper))

	humans.Init(helper)
	humansRouter.Init(api.Group("/humans"), humans.H)

	schedules.Init(helper)
	schedulesRouter.Init(api.Group("/schedules"), schedules.H)

	m2m.Init(helper)
	m2mRouter.Init(api.Group("/m2m"), m2m.H)

	menus.Init(helper)
	menusRouter.Init(api.Group("/menus"), menus.H)

	perfoms.Init(helper)
	perfomsRouter.Init(api.Group("/perfoms"), perfoms.H)

	speakers.Init(helper)
	speakersRouter.Init(api.Group("/speakers"), speakers.H)

	eventdays.Init(helper)
	eventdaysRouter.Init(api.Group("/event-days"), eventdays.H)

	place.Init(helper)
	placeRouter.Init(api.Group("/places"), place.H)

	experiences.Init(helper)
	experiencesRouter.Init(api.Group("/experiences"), experiences.H)

	customsections.Init(helper)
	customsectionsRouter.Init(api.Group("/custom-sections"), customsections.H)

	eventmessages.Init(helper)
	eventMessagesRouter.Init(api.Group("/event-messages"), eventmessages.H)

	usersEventsActivitiesRouter.Init(api.Group("/users-events-activities"), userseventssactivities.CreateHandler(helper))

	users.Init(helper)
	usersRouter.Init(api.Group("/users"), users.H)

	searchRouter.Init(api.Group("/search"), search.CreateHandler(helper))
	//
	contacts.Init(helper)
	contactsRouter.Init(api.Group("/contacts"), contacts.H)

	emails.Init(helper)
	emailsRouter.Init(api.Group("/emails"), emails.H)

	phones.Init(helper)
	phonesRouter.Init(api.Group("/phones"), phones.H)

	sessions.Init(helper)
	sessionsRouter.Init(api.Group("/sessions"), sessions.H)
}
