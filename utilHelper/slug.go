package utilHelper

import (
	"github.com/gosimple/slug"
)

func (h *UtilHelper) MakeSlug(forSlug string) string {
	return slug.Make(forSlug)
}
