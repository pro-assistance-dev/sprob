package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ColorTheme struct {
	bun.BaseModel `bun:"color_themes,alias:color_themes"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	Label         string        `json:"label"`
}

type ColorThemes []*ColorTheme

type ColorThemesWithCount struct {
	ColorThemes ColorThemes `json:"items"`
	Count       int         `json:"count"`
}
