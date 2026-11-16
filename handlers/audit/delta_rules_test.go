package audit

import "testing"

// Правило дельты по операции — ТОЙ ЖЕ функцией, что и в `handler.go`.
//
// ⚠️ Раньше тест подставлял заглушку (`switch` в самом тесте) и «проверял сам
// себя». Правило вынесено в `deltaForOperation`, и тест держит именно
// ПРОДАКШН-код: создание и удаление — без перечисления полей, правка — с ним.
func TestDeltaForOperation(t *testing.T) {
	oldRaw := `{"id":"x","name":"Старое","count":1}`
	createRaw := `{"id":"x","name":"Новое","count":2,"contact":{"a":1}}`

	if d := deltaForOperation("create", "", createRaw, false); len(d) != 0 {
		t.Errorf("create дал %d полей дельты, хотим 0: %+v", len(d), d)
	}
	if d := deltaForOperation("update", oldRaw, `{"name":"Новое","count":1}`, false); len(d) != 1 {
		t.Errorf("update должен дать 1 поле, получили %d: %+v", len(d), d)
	}
	// Удаление: содержательное — что было, но без `id`.
	d := deltaForOperation("delete", oldRaw, "", false)
	if len(d) != 2 {
		t.Errorf("delete должен показать 2 поля, получили %d: %+v", len(d), d)
	}
	for _, item := range d {
		if item.Field == "id" {
			t.Error("id попал в дельту удаления")
		}
	}
	// Откат — всегда без перечисления.
	if d := deltaForOperation("update", oldRaw, createRaw, true); len(d) != 0 {
		t.Errorf("откат дал %d полей, хотим 0", len(d))
	}
}
