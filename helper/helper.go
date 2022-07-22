package helper

import (
	"flag"
	"fmt"
	"github.com/pro-assistance/pro-assister/broker"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/db"
	"github.com/pro-assistance/pro-assister/elasticSearchHelper"
	"github.com/pro-assistance/pro-assister/emailHelper"
	"github.com/pro-assistance/pro-assister/httpHelper"
	"github.com/pro-assistance/pro-assister/pdfHelper"
	"github.com/pro-assistance/pro-assister/search"
	"github.com/pro-assistance/pro-assister/socialHelper"
	"github.com/pro-assistance/pro-assister/sqlHelper"
	"github.com/pro-assistance/pro-assister/templater"
	"github.com/pro-assistance/pro-assister/tokenHelper"
	"github.com/pro-assistance/pro-assister/uploadHelper"
	"github.com/pro-assistance/pro-assister/utilHelper"
	"github.com/uptrace/bun/migrate"
	"net/http"
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
}

func NewHelper(config config.Config) *Helper {
	http := httpHelper.NewHTTPHelper(config)
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
	return &Helper{HTTP: http, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: social, Search: search, Util: util, Templater: templ, Broker: brok, DB: dbHelper}
}

func (i *Helper) Run(migrations *migrate.Migrations, handler http.Handler) {
	mode := flag.String("mode", "run", "init/create")
	action := flag.String("action", "migrate", "init/create/createSql/run/rollback")
	name := flag.String("name", "dummy", "init/create/createSql/run/rollback")
	flag.Parse()
	search.InitSearchGroupsTables(i.DB.DB)
	i.DB.DoAction(migrations, name, action)
	if Mode(*mode) == Migrate {
		return
	}
	if Mode(*mode) == Dump {
		fmt.Println("DUMP!")
		i.DB.Dump()
		return
	}
	defer i.DB.DB.Close()
	i.HTTP.ListenAndServe(handler)
}
