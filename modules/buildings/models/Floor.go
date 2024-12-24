package models

import "github.com/google/uuid"

type Floor struct {
	ID         uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	BuildingID uuid.UUID     `bun:"type:uuid" json:"buildingId"`
	Number     int           `bun:"type:integer" json:"number"`
}

type Floors []*Floor

type FloorsWithCount struct {
	Floors Floors `json:"items"`
	Count  int    `json:"count"`
}
