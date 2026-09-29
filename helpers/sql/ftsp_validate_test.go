package sql

import (
	"errors"
	"net/http"
	"testing"

	httperr "github.com/pro-assistance-dev/sprob/helpers/http"
	"github.com/pro-assistance-dev/sprob/helpers/project"
	"github.com/pro-assistance-dev/sprob/helpers/sql/filter"
	"github.com/pro-assistance-dev/sprob/helpers/sql/sorter"
)

// Ф4.3: неизвестная модель/поле в FTSP — ошибка ЗАПРОСА (400), а не паника 500.
//
// ⚠️ ЧТО ЗАЩИЩАЕМ. Имена полей клиент передаёт строками; раньше неизвестное имя
// доезжало до резолва в схеме и падало nil-pointer'ом — наружу уходил 500
// «ошибка на сервере». Из-за этого было не понять, что дело в опечатке клиента
// (ровно этот класс давал «фильтр молча не работает» в rdkb-map).

func testSchemas() {
	room := &project.Schema{
		NameTable:  "rooms",
		NamePascal: "Room",
		NameCamel:  "room",
		FieldsMap: map[string]*project.SchemaField{
			"name":     {NamePascal: "Name", NameCamel: "name", NameCol: "name"},
			"fullCode": {NamePascal: "FullCode", NameCamel: "fullCode", NameCol: "full_code"},
		},
	}
	building := &project.Schema{
		NameTable:  "buildings",
		NamePascal: "Building",
		NameCamel:  "building",
		FieldsMap: map[string]*project.SchemaField{
			"name": {NamePascal: "Name", NameCamel: "name", NameCol: "name"},
		},
	}
	project.SchemasLib = project.SchemasMap{"room": room, "building": building}
}

func TestFTSPValidate_Ok(t *testing.T) {
	testSchemas()
	ftsp := FTSP{
		F: filter.FilterModels{{
			Model: "room", Col: "fullCode", Type: filter.StringType, Operator: filter.Eq,
		}},
		S: sorter.SortModels{{Model: "room", Col: "name"}},
	}
	if err := ftsp.Validate(); err != nil {
		t.Fatalf("корректный FTSP не должен падать: %v", err)
	}
}

func TestFTSPValidate_UnknownField(t *testing.T) {
	testSchemas()
	ftsp := FTSP{
		F: filter.FilterModels{{
			Model: "room", Col: "нетТакогоПоля", Type: filter.StringType, Operator: filter.Eq,
		}},
	}

	err := ftsp.Validate()

	var unknown ErrUnknownField
	if !errors.As(err, &unknown) {
		t.Fatalf("ожидали ErrUnknownField, получили %v", err)
	}
	if unknown.Field != "нетТакогоПоля" || unknown.Model != "room" {
		t.Fatalf("неверные детали ошибки: %+v", unknown)
	}
	// Текст должен называть проблему, чтобы клиент понял без чтения кода.
	if got := err.Error(); got != `неизвестное поле "нетТакогоПоля" у модели "room"` {
		t.Fatalf("текст ошибки: %q", got)
	}
}

func TestFTSPValidate_UnknownModel(t *testing.T) {
	testSchemas()
	ftsp := FTSP{
		F: filter.FilterModels{{Model: "нетТакойМодели", Col: "name"}},
	}

	err := ftsp.Validate()

	var unknown ErrUnknownField
	if !errors.As(err, &unknown) {
		t.Fatalf("ожидали ErrUnknownField, получили %v", err)
	}
	if unknown.Model != "нетТакойМодели" {
		t.Fatalf("модель в ошибке: %+v", unknown)
	}
}

func TestFTSPValidate_UnknownSortField(t *testing.T) {
	testSchemas()
	ftsp := FTSP{S: sorter.SortModels{{Model: "room", Col: "нетКолонки"}}}

	err := ftsp.Validate()

	var unknown ErrUnknownField
	if !errors.As(err, &unknown) {
		t.Fatalf("ожидали ErrUnknownField, получили %v", err)
	}
	if unknown.Field != "нетКолонки" {
		t.Fatalf("поле в ошибке: %+v", unknown)
	}
}

// Join-фильтр адресуется к колонке ПРИСОЕДИНЯЕМОЙ модели — её и проверяем.
func TestFTSPValidate_JoinUsesJoinModel(t *testing.T) {
	testSchemas()
	ftsp := FTSP{
		F: filter.FilterModels{{
			Model: "room", JoinTableModel: "building", Col: "name",
			Type: filter.JoinType, Operator: filter.Eq,
		}},
	}
	if err := ftsp.Validate(); err != nil {
		t.Fatalf("join по существующей колонке не должен падать: %v", err)
	}

	bad := FTSP{
		F: filter.FilterModels{{
			Model: "room", JoinTableModel: "building", Col: "fullCode",
			Type: filter.JoinType, Operator: filter.Eq,
		}},
	}
	err := bad.Validate()
	var unknown ErrUnknownField
	if !errors.As(err, &unknown) {
		t.Fatalf("опечатка в join-колонке должна ловиться: %v", err)
	}
	if unknown.Model != "building" {
		t.Fatalf("должна называться ПРИСОЕДИНЯЕМАЯ модель: %+v", unknown)
	}
}

// Ошибка валидации отдаётся клиенту как 400 (контракт с `helpers/http`).
func TestErrUnknownField_StatusIsBadRequest(t *testing.T) {
	err := ErrUnknownField{Model: "room", Field: "nope"}

	if status := httperr.StatusForError(err); status != http.StatusBadRequest {
		t.Fatalf("ожидали 400, получили %d", status)
	}
}

// Прочие ошибки не должны превратиться в 400 случайно.
func TestStatusForError_OtherStays500(t *testing.T) {
	if status := httperr.StatusForError(errors.New("что-то сломалось")); status != http.StatusInternalServerError {
		t.Fatalf("обычная ошибка должна остаться 500, получили %d", status)
	}
}
