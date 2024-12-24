package models

import (
	"github.com/google/uuid"
	baseModels "github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/uptrace/bun"
)

type ExtractField struct {
	bun.BaseModel `bun:"extract_fields,alias:extract_fields"`
	ID            uuid.NullUUID           `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	SchemaFieldID uuid.NullUUID           `bun:"type:uuid" json:"schemaFieldId"`
	SchemaField   *baseModels.SchemaField `bun:"rel:belongs-to" json:"schemaField"`
}

type ExtractFields []*ExtractField

type ExtractFieldsWithCount struct {
	ExtractFields ExtractFields `json:"items"`
	Count         int           `json:"count"`
}
