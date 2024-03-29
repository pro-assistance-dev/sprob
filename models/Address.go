package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Address struct {
	bun.BaseModel `bun:"address,alias:address"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Region        string        `json:"region"`
	City          string        `json:"city"`
	Street        string        `json:"street"`
	Building      string        `json:"building"`
	Flat          string        `json:"flat"`
	Zip           int           `json:"zip"`

	RegionID   string `json:"regionId"`
	CityID     string `json:"cityId"`
	StreetID   string `json:"streetId"`
	BuildingID string `bun:"b_id" json:"buildingId"`

	Contact   *Contact      `bun:"rel:belongs-to" json:"contact"`
	ContactID uuid.NullUUID `bun:"cii,type:uuid" json:"contactId"`
}

type Addresss []*Address

func (item *Address) GetFullAddress() string {
	return fmt.Sprintf("%d, %s, %s, %s, %s", item.Zip, item.Region, item.City, item.Street, item.Building)
}
