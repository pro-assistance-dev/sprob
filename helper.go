package main

import (
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/elasticSearchHelper"
	"github.com/pro-assistance/pro-assister/emailHelper"
	"github.com/pro-assistance/pro-assister/httpHelper"
	"github.com/pro-assistance/pro-assister/pdfHelper"
	"github.com/pro-assistance/pro-assister/socialHelper"
	"github.com/pro-assistance/pro-assister/sqlHelper"
	"github.com/pro-assistance/pro-assister/tokenHelper"
	"github.com/pro-assistance/pro-assister/uploadHelper"
)

type Helper struct {
	HTTP     *httpHelper.HTTPHelper
	Search   *elasticSearchHelper.ElasticSearchHelper
	PDF      *pdfHelper.PDFHelper
	Uploader uploadHelper.Uploader
	SQL      *sqlHelper.SQLHelper
	Token    *tokenHelper.TokenHelper
	Email    *emailHelper.EmailHelper
	Social   *socialHelper.SocialHelper
}

func NewHelper(config config.Config) *Helper {
	http := httpHelper.NewHTTPHelper()
	pdf := pdfHelper.NewPDFHelper(config)
	sql := sqlHelper.NewSQLHelper()
	uploader := uploadHelper.NewLocalUploader(&config.UploadPath)
	token := tokenHelper.NewTokenHelper(config.TokenSecret)
	email := emailHelper.NewEmailHelper(config.Email)
	social := socialHelper.NewSocial(config.Social)
	search := elasticSearchHelper.NewElasticSearchHelper(config.ElasticSearch.ElasticSearchOn)
	return &Helper{HTTP: http, Uploader: uploader, PDF: pdf, SQL: sql, Token: token, Email: email, Social: social, Search: search}
}
