package models

import (
	"github.com/google/uuid"
)

type PostAddress struct {
	ID          uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Address     string        `json:"address"`
	Description string        `json:"description"`
	Contact     *Contact      `bun:"rel:belongs-to" json:"contact"`
	ContactID   uuid.NullUUID `bun:"type:uuid" json:"contactId"`
	Main        bool          `json:"main"`
}

type PostAddresses []*PostAddress

type PostAddressWithCount struct {
	PostAddresse PostAddresses `json:"items"`
	Count        int           `json:"count"`
}
