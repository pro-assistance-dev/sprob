package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ChatMessage[UserT any] struct {
	bun.BaseModel `bun:"chat_messages,alias:chat_messages"`
	ID            uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	UserID        uuid.NullUUID   `bun:"type:uuid" json:"userId"`
	User          UserT           `bun:"rel:belongs-to" json:"user"`
	Chat          *Chat[UserT]    `bun:"rel:belongs-to" json:"chat"`
	ChatID        uuid.NullUUID   `bun:"type:uuid" json:"chatId"`
	Message       string          `json:"message"`
	Type          ChatMessageType `bun:"-" json:"type"`
	CreatedOn     time.Time       `bun:",nullzero,notnull" json:"createdOn"`
}

type ChatMessageType string

const (
	ping    ChatMessageType = "ping"
	join    ChatMessageType = "join"
	exit    ChatMessageType = "exit"
	message ChatMessageType = "message"
	// write   ChatMessageType = "write"
)

type ChatMessages[UserT any] []*ChatMessage[UserT]

func (item *ChatMessage[UserT]) IsPing() bool {
	return item.Type == ping
}

func (item *ChatMessage[UserT]) ToBytes() []byte {
	b, _ := json.Marshal(item)
	return b
}

type ChatMessagesWithCount[UserT any] struct {
	ChatMessages ChatMessages[UserT] `json:"items"`
	Count        int                 `json:"count"`
}
