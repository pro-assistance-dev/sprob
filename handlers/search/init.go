package search

import (
	"github.com/pro-assistance/pro-assister/helper"
	"github.com/pro-assistance/pro-assister/models"

	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Search(c *gin.Context)
}

type IService interface {
	Search(*models.SearchModel) error
}

type IRepository interface {
	GetGroupByKey(string) (*models.SearchGroup, error)
	Search(*models.SearchModel) error
}

type Handler struct {
	helper *helper.Helper
}

type Service struct {
	helper *helper.Helper
}

type Repository struct {
	helper *helper.Helper
}

var (
	H *Handler
	S *Service
	R *Repository
)

func Init(h *helper.Helper) {
	H = &Handler{helper: h}
	S = &Service{helper: h}
	R = &Repository{helper: h}
}
