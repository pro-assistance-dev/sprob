package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Contact struct {
	bun.BaseModel `bun:"contacts,alias:contacts"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Time          string        `json:"time"`
	Description   string        `json:"description"`
	Latitude      string        `json:"latitude"`
	Longitude     string        `json:"longitude"`
	Emails        Emails        `bun:"rel:has-many" json:"emails"`
	PostAddresses PostAddresses `bun:"rel:has-many" json:"postAddresses"`
	Phones        Phones        `bun:"rel:has-many" json:"phones"`
	Websites      Websites      `bun:"rel:has-many" json:"websites"`
	Address       *Address      `bun:"rel:has-one,join:id=cii" json:"address"`
}

type Contacts []*Contact

type ContactsWithCount struct {
	Contacts Contacts `json:"items"`
	Count    int      `json:"count"`
}

func (item *Contact) SetIDForChildren() {
	for i := range item.Emails {
		item.Emails[i].ContactID = item.ID
	}
	for i := range item.PostAddresses {
		item.PostAddresses[i].ContactID = item.ID
	}
	for i := range item.Phones {
		item.Phones[i].ContactID = item.ID
	}
	for i := range item.Websites {
		item.Websites[i].ContactID = item.ID
	}
	item.Address.ContactID = item.ID
}

func (items Contacts) SetIDForChildren() {
	for i := range items {
		items[i].SetIDForChildren()
	}
}

func (items Contacts) GetEmails() Emails {
	itemsForGet := make(Emails, 0)
	for i := range items {
		itemsForGet = append(itemsForGet, items[i].Emails...)
	}
	return itemsForGet
}

func (items Contacts) GetPostAddresses() PostAddresses {
	itemsForGet := make(PostAddresses, 0)
	for i := range items {
		itemsForGet = append(itemsForGet, items[i].PostAddresses...)
	}
	return itemsForGet
}

func (items Contacts) GetTelephoneNumbers() Phones {
	itemsForGet := make(Phones, 0)
	for i := range items {
		itemsForGet = append(itemsForGet, items[i].Phones...)
	}
	return itemsForGet
}

func (items Contacts) GetWebsites() Websites {
	itemsForGet := make(Websites, 0)
	for i := range items {
		itemsForGet = append(itemsForGet, items[i].Websites...)
	}
	return itemsForGet
}
