package helper

import (
	"flag"
	"log"
	"net/http"

	"github.com/pro-assistance-dev/sprob/config"
	"github.com/pro-assistance-dev/sprob/helpers/broker"
	"github.com/pro-assistance-dev/sprob/helpers/cron"
	"github.com/pro-assistance-dev/sprob/helpers/db"
	"github.com/pro-assistance-dev/sprob/helpers/email"
	"github.com/pro-assistance-dev/sprob/helpers/logger"
	"github.com/pro-assistance-dev/sprob/helpers/pdf"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/search"
	"github.com/pro-assistance-dev/sprob/helpers/social"
	"github.com/pro-assistance-dev/sprob/helpers/sql"
	"github.com/pro-assistance-dev/sprob/helpers/templater"
	"github.com/pro-assistance-dev/sprob/helpers/token"
	"github.com/pro-assistance-dev/sprob/helpers/uploader"
	"github.com/pro-assistance-dev/sprob/helpers/util"
	"github.com/pro-assistance-dev/sprob/helpers/validator"

	httpHelper "github.com/pro-assistance-dev/sprob/helpers/http"

	"github.com/gin-gonic/gin"
	"github.com/oiime/logrusbun"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun/migrate"
)

type Helper struct {
	HTTP      *httpHelper.HTTP
	PDF       *pdf.PDF
	Uploader  uploader.Uploader
	SQL       *sql.SQL
	Token     *token.Token
	Email     *email.Email
	Social    *social.Social
	Util      *util.Util
	Templater *templater.Templater
	Broker    *broker.Broker
	DB        *db.DB
	Validator *validator.Validator
	Cron      *cron.Cron
	Project   *project.Project
	Logger    *logrus.Logger
}

func NewHelper(c config.Config) *Helper {
	h := httpHelper.NewHTTP(c.Server)
	pdf := pdf.NewPDF(c.Project)
	sql := sql.NewSQL()
	uploader := uploader.NewLocalUploader(&c.Project.UploadPath)
	token := token.NewToken(c.Token)
	email := email.NewEmail(c.Email)
	soc := social.NewSocial(c.Social)
	util := util.NewUtil(c.Project.BinPath)
	templ := templater.NewTemplater(c.Project)
	db := db.NewDB(c.DB)
	brok := broker.NewBroker()
	v := validator.NewValidator()
	cr := cron.NewCron()
	ph := project.NewProject(&c.Project)
	l := logger.NewLogger()
	return &Helper{HTTP: h, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: soc, Util: util, Templater: templ, Broker: brok, DB: db, Validator: v, Cron: cr, Project: ph, Logger: l}
}

type RouterHandler interface {
	Use(...gin.HandlerFunc) gin.IRoutes
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func (i *Helper) Run(migrations []*migrate.Migrations, routerInitFunc func(*gin.Engine, *Helper)) Mode {
	mode := flag.String("mode", "run", "init/create")
	action := flag.String("action", "migrate", "init/create/createSql/run/rollback")
	name := flag.String("name", "dummy", "init/create/createSql/run/rollback")
	flag.Parse()

	if Mode(*mode) == Dump {
		err := i.DB.Dump()
		if err != nil {
			log.Fatal(err)
		}
		return Dump
	}
	if Mode(*mode) == Migrate {
		search.InitSearchGroupsTables(i.DB.DB)
		err := i.DB.DoAction(migrations, *name, action)
		if err != nil {
			log.Fatal(err)
		}
		return Migrate
	}

	i.DB.DB.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: i.Logger, ErrorLevel: logrus.ErrorLevel, QueryLevel: logrus.DebugLevel}))

	defer i.DB.DB.Close()
	i.Project.InitSchemas()
	search.InitSearchGroupsTables(i.DB.DB)

	i.HTTP.ListenAndServe(initRouter(i, routerInitFunc))
	return Listen
}

func initRouter(h *Helper, routerInitFunc func(*gin.Engine, *Helper)) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.LoggingMiddleware(h.Logger))
	router.Use(logger.LoggingMiddleware(h.Logger), gin.Recovery())
	routerInitFunc(router, h)
	return router
}
