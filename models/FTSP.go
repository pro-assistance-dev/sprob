package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FTSP struct {
	bun.BaseModel `bun:"ftsp,alias:ftsp"`
	ID            uuid.NullUUID `json:"id"`
	FTSP          string        `json:"ftsp"`
	Name          string        `json:"name"`
}
