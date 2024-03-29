package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Phone struct {
	bun.BaseModel `bun:"phones,alias:phones"`
	ID            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Number        string        `json:"number"`
	Description   string        `json:"description"`
	Contact       *Contact      `bun:"rel:belongs-to" json:"contact"`
	ContactID     uuid.NullUUID `bun:"type:uuid" json:"contactId"`
	Main          bool          `json:"main"`
}

type Phones []*Phone

type PhonesWithCount struct {
	Phones Phones `json:"items"`
	Count  int    `json:"count"`
}
