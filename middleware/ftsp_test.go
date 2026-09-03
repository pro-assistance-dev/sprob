package middleware

import (
	"reflect"
	"testing"

	"github.com/pro-assistance-dev/sprob/helpers/sql"
)

// Т7.2: FTSPStore живёт в Middleware (не в пакетном глобале) — состояние
// изолировано между экземплярами (тесты/параллельные запросы).
func TestFTSPStore_IsolatedBetweenMiddlewares(t *testing.T) {
	m1 := CreateMiddleware(nil)
	m2 := CreateMiddleware(nil)

	q := &sql.FTSPQuery{FTSP: sql.FTSP{}}
	if _, found := m1.ftsp.GetOrCreateFTSP(q); !found {
		t.Fatalf("первый GetOrCreate должен сохранить запрос (QID=%q)", q.QID)
	}
	if q.QID == "" {
		t.Fatal("GetOrCreateFTSP не выдал QID")
	}

	// Тот же QID в другом Middleware — не найден (изоляция).
	if _, found := m2.ftsp.GetFTSP(q.QID); found {
		t.Error("store второго middleware содержит чужой QID — состояние не изолировано")
	}
	// В своём — найден.
	if _, found := m1.ftsp.GetFTSP(q.QID); !found {
		t.Error("свой QID не найден в своём store")
	}
}

func TestFTSPStore_GetOrCreateByQID(t *testing.T) {
	m := CreateMiddleware(nil)

	// Предзаполненный QID (фронт прислал qid) — возврат по ключу.
	first := &sql.FTSPQuery{FTSP: sql.FTSP{}}
	if _, found := m.ftsp.GetOrCreateFTSP(first); !found {
		t.Fatal("create failed")
	}
	second := &sql.FTSPQuery{QID: first.QID}
	got, found := m.ftsp.GetOrCreateFTSP(second)
	if !found {
		t.Fatal("повторный GetOrCreate по QID не нашёл запись")
	}
	if !reflect.DeepEqual(got, first.FTSP) {
		t.Error("значение по QID не совпадает с сохранённым")
	}
}
