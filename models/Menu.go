package models

import (
	"github.com/pro-assistance-dev/sprob/helpers/uploader"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Menu struct {
	bun.BaseModel `bun:"menus,alias:menus"`
	ID            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	Link          string        `json:"link"`
	Top           bool          `json:"top"`
	Side          bool          `json:"side"`
	Order         uint          `bun:"menu_order" json:"order"`
	PageID        uuid.NullUUID `bun:"type:uuid" json:"pageId"`
	Icon          *FileInfo     `bun:"rel:belongs-to" json:"icon"`
	IconID        uuid.NullUUID `bun:"type:uuid"  json:"iconId"`
	Hide          bool          `json:"hide"`

	SubMenus          SubMenus    `bun:"rel:has-many" json:"subMenus"`
	SubMenusForDelete []uuid.UUID `bun:"-" json:"subMenusForDelete"`
}

type Menus []*Menu

type MenusWithCount struct {
	Menus Menus `json:"items"`
	Count int   `json:"count"`
}

func (items Menus) GetIcons() FileInfos {
	itemsForGet := make(FileInfos, 0)
	for _, item := range items {
		itemsForGet = append(itemsForGet, item.Icon)
	}
	return itemsForGet
}

func (items Menus) GetSubMenus() SubMenus {
	itemsForGet := make(SubMenus, 0)
	for _, item := range items {
		itemsForGet = append(itemsForGet, item.SubMenus...)
	}
	return itemsForGet
}

func (items Menus) GetSubMenusForDelete() []uuid.UUID {
	itemsForGet := make([]uuid.UUID, 0)
	for _, item := range items {
		itemsForGet = append(itemsForGet, item.SubMenusForDelete...)
	}
	return itemsForGet
}

func (item *Menu) SetIDForChildren() {
	if len(item.SubMenus) == 0 {
		return
	}
	for i := range item.SubMenus {
		item.SubMenus[i].MenuID = item.ID
	}
}

func (items Menus) SetIDForChildren() {
	for i := range items {
		items[i].SetIDForChildren()
	}
}

func (item *Menu) SetFilePath(fileID *string) *string {
	if item.Icon.ID.UUID.String() == *fileID {
		item.Icon.FileSystemPath = uploader.BuildPath(fileID)
		return &item.Icon.FileSystemPath
	}
	path := item.SubMenus.SetFilePath(fileID)
	if path != nil {
		return path
	}
	return nil
}

func (item *Menu) SetForeignKeys() {
	item.IconID = item.Icon.ID
}

func (items Menus) SetForeignKeys() {
	for i := range items {
		items[i].SetForeignKeys()
	}
}
