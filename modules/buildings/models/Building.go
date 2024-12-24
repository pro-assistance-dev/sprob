package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Building struct {
	bun.BaseModel `bun:"buildings"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Number        string    `json:"number"`
	Floors        Floors    `bun:"rel:has-many" json:"floors"`
	Entrances     Entrances `bun:"rel:has-many" json:"entrances"`
	MapNodeName   string    `bun:"-"`
}

type Buildings []*Building

type BuildingsWithCount struct {
	Buildings Buildings `json:"items"`
	Count     int       `json:"count"`
}
