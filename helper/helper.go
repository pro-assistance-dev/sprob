package helper

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oiime/logrusbun"
	"github.com/pro-assistance/pro-assister/broker"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/cronHelper"
	"github.com/pro-assistance/pro-assister/db"
	"github.com/pro-assistance/pro-assister/elasticSearchHelper"
	"github.com/pro-assistance/pro-assister/emailHelper"
	"github.com/pro-assistance/pro-assister/httpHelper"
	"github.com/pro-assistance/pro-assister/loggerhelper"
	"github.com/pro-assistance/pro-assister/pdfHelper"
	"github.com/pro-assistance/pro-assister/projecthelper"
	"github.com/pro-assistance/pro-assister/search"
	"github.com/pro-assistance/pro-assister/socialHelper"
	"github.com/pro-assistance/pro-assister/sqlHelper"
	"github.com/pro-assistance/pro-assister/templater"
	"github.com/pro-assistance/pro-assister/tokenHelper"
	"github.com/pro-assistance/pro-assister/uploadHelper"
	"github.com/pro-assistance/pro-assister/utilHelper"
	"github.com/pro-assistance/pro-assister/validatorhelper"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun/migrate"
)

type Helper struct {
	HTTP      *httpHelper.HTTPHelper
	Search    *elasticSearchHelper.ElasticSearchHelper
	PDF       *pdfHelper.PDFHelper
	Uploader  uploadHelper.Uploader
	SQL       *sqlHelper.SQLHelper
	Token     *tokenHelper.TokenHelper
	Email     *emailHelper.EmailHelper
	Social    *socialHelper.SocialHelper
	Util      *utilHelper.UtilHelper
	Templater *templater.Templater
	Broker    *broker.Broker
	DB        *db.DB
	Validator *validatorhelper.Validator
	Cron      *cronHelper.Cron
	Project   *projecthelper.ProjectHelper
	Logger    *logrus.Logger
}

func NewHelper(config config.Config) *Helper {
	http := httpHelper.NewHTTPHelper(config.Server)
	pdf := pdfHelper.NewPDFHelper(config)
	sql := sqlHelper.NewSQLHelper()
	uploader := uploadHelper.NewLocalUploader(&config.UploadPath)
	token := tokenHelper.NewTokenHelper(config.Token)
	email := emailHelper.NewEmailHelper(config.Email)
	social := socialHelper.NewSocial(config.Social)
	search := elasticSearchHelper.NewElasticSearchHelper(config.ElasticSearch.ElasticSearchOn)
	util := utilHelper.NewUtilHelper(config.BinPath)
	templ := templater.NewTemplater(config)
	dbHelper := db.NewDBHelper(config.DB)
	brok := broker.NewBroker()
	v := validatorhelper.NewValidator()
	cr := cronHelper.NewCronHelper()
	ph := projecthelper.NewProjectHelper()
	l := loggerhelper.NewLogger()
	return &Helper{HTTP: http, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: social, Search: search, Util: util, Templater: templ, Broker: brok, DB: dbHelper, Validator: v, Cron: cr, Project: ph, Logger: l}
}

type RouterHandler interface {
	Use(...gin.HandlerFunc) gin.IRoutes
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func (i *Helper) Run(migrations *migrate.Migrations, handler RouterHandler, init func(http.Handler, *Helper)) Mode {
	mode := flag.String("mode", "run", "init/create")
	action := flag.String("action", "migrate", "init/create/createSql/run/rollback")
	name := flag.String("name", "dummy", "init/create/createSql/run/rollback")
	flag.Parse()
	if Mode(*mode) == Run {
		handler.Use(loggerhelper.LoggingMiddleware(i.Logger), gin.Recovery())
		i.DB.DB.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: i.Logger, ErrorLevel: logrus.ErrorLevel, QueryLevel: logrus.DebugLevel}))
		init(handler, i)
		return Run
	}
	if Mode(*mode) == Dump {
		i.DB.Dump()
		return Dump
	}
	if Mode(*mode) == Migrate {
		search.InitSearchGroupsTables(i.DB.DB)
		i.DB.DoAction(migrations, name, action)
		return Migrate
	}
	defer i.DB.DB.Close()
	i.Project.InitSchemas()
	search.InitSearchGroupsTables(i.DB.DB)
	i.DB.DoAction(migrations, name, action)

	i.HTTP.ListenAndServe(handler)
	return Listen
}
