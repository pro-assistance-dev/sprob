package contacts

import (
	"github.com/pro-assistance-dev/sprob/helper"
)

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
