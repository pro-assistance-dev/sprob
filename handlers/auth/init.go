package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/models"
)

type IHandler interface {
	CheckUUID(c *gin.Context)
	RefreshPassword(c *gin.Context)
}

type IService interface {
	Register(string, string) (string, error)
	Login(*models.UserAccount) (string, error)
	CheckUUID(context.Context, string, uuid.NullUUID)
	RefreshPassword(c context.Context)
}

type IRepository interface {
	Create(context.Context, *models.UserAccount) error
	GetByEmail(context.Context, string) (*models.UserAccount, error)
	GetByUUID(context.Context, string) (*models.UserAccount, error)
	UpdateUUID(context.Context, string) error
	UpdatePassword(context.Context, string, string) error
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
