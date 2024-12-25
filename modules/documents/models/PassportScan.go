package models

import (
	"github.com/google/uuid"
	baseModels "github.com/pro-assistance-dev/sprob/models"
	"github.com/uptrace/bun"
)

type PassportScan struct {
	bun.BaseModel `bun:"passport_scans,alias:passport_scans"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `

	Passport   *Passport     `bun:"rel:belongs-to" json:"passport"`
	PassportID uuid.NullUUID `json:"passportId"`

	FileInfo   *baseModels.FileInfo `json:"fileInfo"`
	FileInfoID uuid.NullUUID        `json:"fileInfoId"`
	Order      uint                 `bun:"item_order" json:"order"`
}

type PassportScans []*PassportScan

type PassportScansWithCount struct {
	PassportScans PassportScans `json:"items"`
	Count         int           `json:"count"`
}
