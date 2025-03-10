package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Chat[UserT any] struct {
	bun.BaseModel `bun:"chats"`
	ID            uuid.NullUUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string              `json:"name"`
	CreatedAt     time.Time           `bun:",nullzero,notnull" json:"createdAt"`
	ChatMessages  ChatMessages[UserT] `bun:"rel:has-many,join:id=chat_id" json:"chatMessages"`
}

type Chats[UserT any] []*Chat[UserT]

type ChatsWithCount[UserT any] struct {
	Chats Chats[UserT] `json:"items"`
	Count int          `json:"count"`
}
