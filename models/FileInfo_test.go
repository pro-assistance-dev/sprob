package models

import (
	"encoding/json"
	"testing"
)

// Клиент (sprof) шлёт `itemId` пустой строкой, когда владельца нет.
// Раньше разбор падал на «invalid UUID length: 0» и терялся ВЕСЬ остаток
// тела — включая `id` файла (upload уходил пустым, 500 на загрузке).
func TestFileInfoUnmarshalEmptyUUID(t *testing.T) {
	const body = `{"name":"","url":"x","type":"diet","id":"e28b6af7-60e1-420c-831d-000da4671e8e","originalName":"a.png","fileSystemPath":"1","itemId":"","description":"","order":0}`
	var item FileInfo
	if err := json.Unmarshal([]byte(body), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !item.ID.Valid || item.ID.UUID.String() != "e28b6af7-60e1-420c-831d-000da4671e8e" {
		t.Fatalf("id lost: %+v", item.ID)
	}
	if item.ItemID.Valid {
		t.Fatalf("empty itemId must be NULL, got %+v", item.ItemID)
	}
	if item.Type != "diet" || item.OriginalName != "a.png" {
		t.Fatalf("fields after empty UUID lost: %+v", item)
	}
}

func TestFileInfoUnmarshalOwner(t *testing.T) {
	const body = `{"id":"e28b6af7-60e1-420c-831d-000da4671e8e","itemId":"116833e8-e147-4ec2-841f-f39c304e590f"}`
	var item FileInfo
	if err := json.Unmarshal([]byte(body), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !item.ItemID.Valid || item.ItemID.UUID.String() != "116833e8-e147-4ec2-841f-f39c304e590f" {
		t.Fatalf("itemId lost: %+v", item.ItemID)
	}
}
