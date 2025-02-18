package basehandler

import (
	"github.com/pro-assistance-dev/sprob/helper"
)

type Handler struct {
	helper *helper.Helper
}

type Service struct {
	helper *helper.Helper
}

type Repository[TSingle, TPlural, TPluralWithCount any] struct {
	helper *helper.Helper
}

type FilesService struct {
	helper *helper.Helper
}

// var (
// 	H *Handler
// 	S *Service
// 	R *Repository
// 	F *FilesService
// )

// func Init(h *helper.Helper) {
// 	H = &Handler{helper: h}
// 	S = &Service{helper: h}
// 	R = &Repository{helper: h}
// 	F = &FilesService{helper: h}
// }
