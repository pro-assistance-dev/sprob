package fileinfos

import (
	"github.com/pro-assistance/pro-assister/helper"

	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Download(c *gin.Context)
	Create(c *gin.Context)
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

type FilesService struct {
	helper *helper.Helper
}

var (
	H *Handler
	S *Service
	R *Repository
	F *FilesService
)

func Init(h *helper.Helper) {
	H = &Handler{helper: h}
	S = &Service{helper: h}
	R = &Repository{helper: h}
	F = &FilesService{helper: h}
}
