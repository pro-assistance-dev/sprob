package models

import (
	"github.com/google/uuid"
	baseModels "github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
)

type Inn struct {
	bun.BaseModel `bun:"inns,alias:inns"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Num           string        `json:"num"`

	FileInfo   *baseModels.FileInfo `bun:"rel:belongs-to" json:"fileInfo"`
	FileInfoID uuid.NullUUID        `json:"fileInfoId"`
}

type Inns []*Inn

type InnsWithCount struct {
	Inns  Inns `json:"items"`
	Count int  `json:"count"`
}
