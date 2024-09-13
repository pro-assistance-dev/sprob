package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Email struct {
	bun.BaseModel `bun:"emails,alias:emails"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Address       string        `json:"address"`
	Description   string        `json:"description"`
	ContactID     uuid.NullUUID `bun:"type:uuid" json:"contactId"`
	Main          bool          `json:"main"`
}

type Emails []*Email
