package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Passport struct {
	bun.BaseModel `bun:"passports,alias:passports"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `

	Num          string    `json:"num"`
	Seria        string    `json:"seria"`
	Division     string    `json:"division"`
	DivisionCode string    `json:"divisionCode"`
	Citzenship   string    `json:"citzenship"`
	Date         time.Time `bun:"item_date" json:"date"`

	PassportScans PassportScans `bun:"rel:has-many" json:"passportScans"`
}

type Passports []*Passport

type PassportsWithCount struct {
	Passports Passports `json:"items"`
	Count     int       `json:"count"`
}
