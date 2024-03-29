package models

import (
	"github.com/google/uuid"
)

type Website struct {
	ID          uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Address     string        `json:"address"`
	Description string        `json:"description"`
	Contact     *Contact      `bun:"rel:belongs-to" json:"contact"`
	ContactID   uuid.NullUUID `bun:"type:uuid" json:"contactId"`
	Main        bool          `json:"main"`
}

type Websites []*Website

type WebsitesWithCount struct {
	Websites Websites `json:"items"`
	Count    int      `json:"count"`
}
