package access

import (
	"os"
	"testing"
)

// Реестр сущностей для тестов middleware (в бою реестр наполняет проект через Register).
// Поля повторяют testMatrix из jwks_test.go (rooms / room-engineerings) + buildings.
func TestMain(m *testing.M) {
	Register([]*Entity{
		{
			Key: "rooms", Table: "rooms",
			Fields: map[string]FieldInfo{
				"id":         {JSON: "id", Owner: "R00_ADMIN"},
				"name":       {JSON: "name", Owner: "R01_HOZ"},
				"actualName": {JSON: "actualName", Owner: "R01_HOZ"},
				"area":       {JSON: "area", Owner: "R00_ADMIN"},
			},
		},
		{
			Key: "room-engineerings", Table: "room_engineering", EntityLevel: true,
			Fields: map[string]FieldInfo{
				"electricity": {JSON: "electricity", Owner: "R02_EXPL"},
			},
		},
		{Key: "buildings", Table: "buildings", EntityLevel: true},
	})
	os.Exit(m.Run())
}
