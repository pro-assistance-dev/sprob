package helper

import (
	"flag"
	"net/http"
	"pro-assister/config"
	"pro-assister/helpers/broker"
	"pro-assister/helpers/cron"
	"pro-assister/helpers/db"
	"pro-assister/helpers/email"
	"pro-assister/helpers/logger"
	"pro-assister/helpers/pdf"
	"pro-assister/helpers/project"
	"pro-assister/helpers/search"
	"pro-assister/helpers/social"
	"pro-assister/helpers/sql"
	"pro-assister/helpers/templater"
	"pro-assister/helpers/token"
	"pro-assister/helpers/uploader"
	"pro-assister/helpers/util"
	"pro-assister/helpers/validator"

	httpHelper "pro-assister/helpers/http"

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

func NewHelper(config config.Config) *Helper {
	h := httpHelper.NewHTTP(config.Server)
	pdf := pdf.NewPDF(config)
	sql := sql.NewSQL()
	uploader := uploader.NewLocalUploader(&config.UploadPath)
	token := token.NewToken(config.Token)
	email := email.NewEmail(config.Email)
	soc := social.NewSocial(config.Social)
	util := util.NewUtil(config.BinPath)
	templ := templater.NewTemplater(config)
	db := db.NewDB(config.DB)
	brok := broker.NewBroker()
	v := validator.NewValidator()
	cr := cron.NewCron()
	ph := project.NewProject()
	l := logger.NewLogger()
	return &Helper{HTTP: h, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: soc, Util: util, Templater: templ, Broker: brok, DB: db, Validator: v, Cron: cr, Project: ph, Logger: l}
}

type RouterHandler interface {
	Use(...gin.HandlerFunc) gin.IRoutes
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func (i *Helper) Run(migrations *migrate.Migrations, init func(*gin.Engine, *Helper)) Mode {
	mode := flag.String("mode", "run", "init/create")
	action := flag.String("action", "migrate", "init/create/createSql/run/rollback")
	name := flag.String("name", "dummy", "init/create/createSql/run/rollback")
	flag.Parse()

	if Mode(*mode) == Dump {
		i.DB.Dump()
		return Dump
	}
	if Mode(*mode) == Migrate {
		search.InitSearchGroupsTables(i.DB.DB)
		i.DB.DoAction(migrations, name, action)
		return Migrate
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.LoggingMiddleware(i.Logger))
	router.Use(logger.LoggingMiddleware(i.Logger), gin.Recovery())
	i.DB.DB.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: i.Logger, ErrorLevel: logrus.ErrorLevel, QueryLevel: logrus.DebugLevel}))
	init(router, i)

	defer i.DB.DB.Close()
	i.Project.InitSchemas()
	search.InitSearchGroupsTables(i.DB.DB)
	i.DB.DoAction(migrations, name, action)

	i.HTTP.ListenAndServe(router)
	return Listen
}
