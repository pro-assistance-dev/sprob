package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Email struct {
	bun.BaseModel `bun:"emails,alias:emails"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Address       string        `json:"address"`
	ContactID     uuid.NullUUID `bun:"type:uuid" json:"contactId"`
	Main          bool          `json:"main"`
	EmailMessages EmailMessages `bun:"rel:has-many" json:"emailMessages"`
}

type Emails []*Email
