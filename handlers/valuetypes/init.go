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
	s      *Service
}

type Service struct {
	helper *helper.Helper
	r      *Repository
}

type Repository struct {
	helper *helper.Helper
}

var (
	H *Handler
	S *Service
	R *Repository
)

func Init(h *helper.Helper) *Handler {
	r := &Repository{helper: h}
	s := &Service{helper: h, r: r}
	handler := &Handler{helper: h, s: s}
	H = handler
	S = s
	R = r
	return handler
}
