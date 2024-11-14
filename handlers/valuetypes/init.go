package valuetypes

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/models"
)

type IHandler interface {
	GetAll(c *gin.Context)
	Get(c *gin.Context)
}

type IService interface {
	GetAll(context.Context) (models.ValueTypes, error)
	Get(context.Context, string) (*models.ValueType, error)
}

type IRepository interface {
	GetAll(context.Context) (string, error)
	Get(context.Context, string) (*models.ValueType, error)
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
