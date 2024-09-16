package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Contact struct {
	bun.BaseModel `bun:"contacts,alias:contacts"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
	Emails        Emails        `bun:"rel:has-many" json:"emails"`
	// Email Email `bun:"rel:has-many" json:"email"`
}

type Contacts []*Contact
