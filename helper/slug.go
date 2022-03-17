package main

import "github.com/gosimple/slug"

func (h *Helper) MakeSlug(forSlug string) string {
	return slug.Make(forSlug)
}
