package metabase

import (
	"github.com/pro-assistance-dev/sprob/helper"
)

type Handler struct {
	helper *helper.Helper
}

var H *Handler

func Init(h *helper.Helper) {
	H = &Handler{helper: h}
}
