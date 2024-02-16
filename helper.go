package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oiime/logrusbun"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/helpers/broker"
	"github.com/pro-assistance/pro-assister/helpers/cron"
	"github.com/pro-assistance/pro-assister/helpers/db"
	"github.com/pro-assistance/pro-assister/helpers/email"
	"github.com/pro-assistance/pro-assister/helpers/http"
	"github.com/pro-assistance/pro-assister/helpers/logger"
	"github.com/pro-assistance/pro-assister/helpers/pdf"
	"github.com/pro-assistance/pro-assister/helpers/project"
	"github.com/pro-assistance/pro-assister/helpers/search"
	"github.com/pro-assistance/pro-assister/helpers/social"
	"github.com/pro-assistance/pro-assister/helpers/sql"
	"github.com/pro-assistance/pro-assister/helpers/templater"
	"github.com/pro-assistance/pro-assister/helpers/token"
	"github.com/pro-assistance/pro-assister/helpers/uploader"
	"github.com/pro-assistance/pro-assister/helpers/util"
	"github.com/pro-assistance/pro-assister/helpers/validator"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun/migrate"
)

type Helper struct {
	HTTP      *http.HTTP
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
	http := http.NewHTTP(config.Server)
	pdf := pdf.NewPDF(config)
	sql := sql.NewSQL()
	uploader := uploader.NewLocalUploader(&config.UploadPath)
	token := token.NewToken(config.Token)
	email := email.NewEmail(config.Email)
	social := social.NewSocial(config.Social)
	util := util.NewUtil(config.BinPath)
	templ := templater.NewTemplater(config)
	db := db.NewDB(config.DB)
	brok := broker.NewBroker()
	v := validator.NewValidator()
	cr := cron.NewCron()
	ph := project.NewProject()
	l := logger.NewLogger()
	return &Helper{HTTP: http, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: social, Search: search, Util: util, Templater: templ, Broker: brok, DB: db, Validator: v, Cron: cr, Project: ph, Logger: l}
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