package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Chat struct {
	bun.BaseModel `bun:"chats"`
	ID            uuid.UUID    `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string       `json:"name"`
	CreatedOn     time.Time    `bun:",nullzero,notnull" json:"createdOn"`
	ChatMessages  ChatMessages `bun:"rel:has-many" json:"chatMessages"`
}

type Chats []*Chat

type ChatsWithCount struct {
	Chats Chats `json:"items"`
	Count int   `json:"count"`
}
