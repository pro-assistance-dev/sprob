package fileinfos

import (
	"github.com/pro-assistance-dev/sprob/helper"

	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Download(c *gin.Context)
	Create(c *gin.Context)
}

type Handler struct {
	helper *helper.Helper
	s      *Service
	f      *FilesService
}

type Service struct {
	helper *helper.Helper
	r      *Repository
}

type Repository struct {
	helper *helper.Helper
}

type FilesService struct {
	helper *helper.Helper
}

var (
	H *Handler
	S *Service
	R *Repository
	F *FilesService
)

func Init(h *helper.Helper) *Handler {
	r := &Repository{helper: h}
	s := &Service{helper: h, r: r}
	f := &FilesService{helper: h}
	handler := &Handler{helper: h, s: s}
	handler.f = f
	H = handler
	S = s
	R = r
	F = f
	return handler
}
