package models

import (
	"github.com/google/uuid"
	baseModels "github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
)

type Snils struct {
	bun.BaseModel `bun:"snilss,alias:snilss"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Num           string        `json:"num"`

	FileInfo   *baseModels.FileInfo `json:"fileInfo"`
	FileInfoID uuid.NullUUID        `json:"fileInfoId"`
}

type Snilss []*Snils

type SnilssWithCount struct {
	Snilss Snilss `json:"items"`
	Count  int    `json:"count"`
}
