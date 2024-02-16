package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/helper"
	"github.com/pro-assistance/pro-assister/models"
)

type IService interface {
	Register(email string, password string) (string, error)
	Login(user *models.UserAccount) (string, error)
}

type IRepository interface {
	Create(*gin.Context, *models.UserAccount) error
	GetByEmail(*gin.Context, string) error
}

var (
	S *Service
	R *Repository
)

func Init(h *helper.Helper) {
	R = NewRepository(h)
	S = NewService(h)
}

type Service struct {
	repository IRepository
	helper     *helper.Helper
}

type Repository struct {
	helper *helper.Helper
}

type FilesService struct {
	helper *helper.Helper
}

func NewService(helper *helper.Helper) *Service {
	return &Service{helper: helper}
}

func NewRepository(helper *helper.Helper) *Repository {
	return &Repository{helper: helper}
}
