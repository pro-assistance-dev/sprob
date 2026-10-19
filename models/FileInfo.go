package models

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ⚠️ `item_id`/`item_one_id` — ОБЯЗАТЕЛЬНО `omitempty`: `uuid.NullUUID` маршалится
// объектом `{UUID, Valid}`, и сервер получал `"itemId":{…}` → `invalid UUID
// length: 9` (500 на загрузке фото). Пустой UUID надо не слать вовсе.
type FileInfo struct {
	bun.BaseModel  `bun:"file_infos,alias:file_infos"`
	ID             uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	OriginalName   string        `json:"originalName"`
	FileSystemPath string        `json:"fileSystemPath"`
	ItemID         uuid.NullUUID `bun:"type:uuid" json:"itemId,omitempty"`
	ItemOneID      uuid.NullUUID `bun:"type:uuid" json:"itemOneId,omitempty"`
	Description    string        `json:"description"`
	Order          uint          `bun:"item_order" json:"order"`
	Type           string        `json:"type"`
}

// UnmarshalJSON — пустую строку (`"itemId":""`, `"id":""`) читаем как NULL.
//
// ⚠️ Иначе `uuid.NullUUID.UnmarshalJSON` на `""` сразу возвращает
// `invalid UUID length: 0` и ПРЕРЫВАЕТ разбор всего тела: все поля после
// `itemId` теряются (вместе с `id` файла от AdminGallery). Клиенты шлют пустую
// строку всегда, когда поля нет — это не ошибка клиента.
func (f *FileInfo) UnmarshalJSON(data []byte) error {
	// Копия без uuid-полей: разбор в `uuid.NullUUID` упал бы на `""` раньше,
	// чем мы увидим строку (клиент так делает всегда, когда поля нет).
	plain := struct {
		ID             *string `json:"id"`
		ItemID         *string `json:"itemId"`
		ItemOneID      *string `json:"itemOneId"`
		OriginalName   string  `json:"originalName"`
		FileSystemPath string  `json:"fileSystemPath"`
		Description    string  `json:"description"`
		Order          uint    `json:"order"`
		Type           string  `json:"type"`
	}{}
	if err := json.Unmarshal(data, &plain); err != nil {
		return err
	}
	f.ID = uuidFromString(plain.ID)
	f.ItemID = uuidFromString(plain.ItemID)
	f.ItemOneID = uuidFromString(plain.ItemOneID)
	f.OriginalName = plain.OriginalName
	f.FileSystemPath = plain.FileSystemPath
	f.Description = plain.Description
	f.Order = plain.Order
	f.Type = plain.Type
	return nil
}

// uuidFromString — разобранный uuid или NULL для пустой/битой строки.
func uuidFromString(raw *string) uuid.NullUUID {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return uuid.NullUUID{}
	}
	parsed, err := uuid.Parse(*raw)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: parsed, Valid: true}
}

type FileInfos []*FileInfo

func (item FileInfo) GetOriginalName() string {
	return item.OriginalName
}

func (item FileInfo) GetFullPath() string {
	return item.FileSystemPath
}

func (items FileInfos) GetPathsAndNames() (paths []string, names []string) {
	for _, item := range items {
		paths = append(paths, item.FileSystemPath)
		names = append(names, item.OriginalName)
	}
	return paths, names
}
