package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type EmailMessage struct {
	bun.BaseModel `bun:"email_messages,alias:email_messages"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Name          string        `json:"name"`
}

type EmailMessages []*EmailMessage
