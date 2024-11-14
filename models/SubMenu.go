package models

import (
	"github.com/pro-assistance-dev/sprob/helpers/uploader"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SubMenu struct {
	bun.BaseModel `bun:"sub_menus,alias:sub_menus"`
	ID            uuid.UUID     `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Color         string        `json:"color"`
	Link          string        `json:"link"`
	Menu          *Menu         `bun:"rel:belongs-to" json:"menus"`
	Hide          bool          `json:"hide"`
	MenuID        uuid.UUID     `bun:"type:uuid" json:"menuId"`
	Order         uint          `bun:"sub_menu_order" json:"order"`
	PageID        uuid.NullUUID `bun:"type:uuid" json:"pageID"`
	Icon          *FileInfo     `bun:"rel:belongs-to" json:"icon"`
	IconID        uuid.NullUUID `bun:"type:uuid"  json:"iconId"`
	IconName      string        `json:"iconName"`
	SvgCode       string        `json:"svgCode"`
}

type SubMenus []*SubMenu

func (items SubMenus) SetFilePath(fileID *string) *string {
	for _, item := range items {
		if item.Icon.ID.UUID.String() == *fileID {
			item.Icon.FileSystemPath = uploader.BuildPath(fileID)
			return &item.Icon.FileSystemPath
		}
	}
	return nil
}

func (items SubMenus) SetForeignKeys() {
	for i := range items {
		items[i].IconID.UUID = items[i].Icon.ID.UUID
		items[i].IconID = items[i].Icon.ID
	}
}

func (items SubMenus) GetFileInfos() FileInfos {
	itemsForGet := make(FileInfos, 0)
	for _, item := range items {
		itemsForGet = append(itemsForGet, item.Icon)
	}
	return itemsForGet
}
