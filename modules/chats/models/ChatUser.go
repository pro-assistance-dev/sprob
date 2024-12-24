package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ChatUser struct {
	bun.BaseModel `bun:"chats_users"`
	ID            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	UserID        uuid.NullUUID `bun:"type:uuid" json:"userId"`
	ChatID        uuid.NullUUID `bun:"type:uuid" json:"chatId"`
	Chat          *Chat         `bun:"rel:belongs-to" json:"chat"`
	JoinTime      time.Time     `bun:",nullzero,notnull" json:"joinTime"`
	ExitTime      time.Time     `bun:",nullzero,notnull" json:"exitTime"`
}

type ChatsUsers []*ChatUser

type ChatsUsersWithCount struct {
	ChatsUsers ChatsUsers `json:"items"`
	Count      int        `json:"count"`
}
