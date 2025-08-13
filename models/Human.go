package models

import (
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

type Human struct {
	bun.BaseModel `bun:"humans,alias:humans"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	Surname       string        `json:"surname"`
	Patronymic    string        `json:"patronymic"`
	DateBirth     *time.Time    `json:"dateBirth"`
	IsMale        bool          `json:"isMale"`
	ItemID        uuid.NullUUID `bun:"type:uuid" json:"itemId"`
}

type Humans []*Human

type HumansWithCount struct {
	Humans Humans `json:"items"`
	Count  int    `json:"count"`
}

func (item *Human) GetFullName() string {
	return fmt.Sprintf("%s %s %s", item.Surname, item.Name, item.Patronymic)
}
