package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FTSPPreset struct {
	bun.BaseModel `bun:"ftsp_presets,alias:ftsp_presets"`
	ID            uuid.NullUUID `json:"id"`
	FTSP          string        `json:"ftsp"`
	Name          string        `json:"name"`
}
type FTSPPresets []*FTSPPreset
