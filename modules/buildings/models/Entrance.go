package models

import "github.com/google/uuid"

type Entrance struct {
	ID          uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Building    *Building     `bun:"rel:belongs-to" json:"building"`
	BuildingID  uuid.UUID     `bun:"type:uuid" json:"buildingId"`
	Number      int           `bun:"type:integer" json:"number"`
	MapNodeName string        `bun:"-"`
}

type Entrances []*Entrance

type EntrancesWithCount struct {
	Entrances Entrances `json:"items"`
	Count     int       `json:"count"`
}
