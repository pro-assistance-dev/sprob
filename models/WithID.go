package models

import "github.com/google/uuid"

type WithID struct {
	ID uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
}
