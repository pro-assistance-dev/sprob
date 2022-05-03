package utilHelper

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

func (h *UtilHelper) MakeSlug(forSlug string, unique bool) string {
	s := slug.Make(forSlug)
	if unique {
		s = fmt.Sprintf("%s-%s", s, uuid.New())
	}
	return s
}
