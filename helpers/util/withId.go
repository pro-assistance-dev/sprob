package util

import "github.com/google/uuid"

type WithId struct {
	ID uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
}
