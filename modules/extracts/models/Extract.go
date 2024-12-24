package models

import (
	"github.com/google/uuid"
	baseModels "github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

type Extract struct {
	bun.BaseModel `bun:"extracts,alias:extracts"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`

	SchemaID uuid.NullUUID      `bun:"type:uuid" json:"schemaId"`
	Schema   *baseModels.Schema `bun:"rel:belongs-to" json:"schema"`
}

type Extracts []*Extract

type ExtractsWithCount struct {
	Extracts Extracts `json:"items"`
	Count    int      `json:"count"`
}
