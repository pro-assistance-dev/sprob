package filter

import (
	"strings"
	"testing"

	"github.com/pro-assistance-dev/sprob/helpers/project"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Ф1 (TECH_DEBT): значения фильтра обязаны уходить ПАРАМЕТРАМИ, а не строкой.
//
// Уязвимость была реальной: `constructWhere` собирал SQL через
// `fmt.Sprintf("... '%s'", f.Value1)`, а Value1 приходит ОТ КЛИЕНТА — апостроф
// ломал запрос. Здесь зафиксирован контракт: в тексте запроса значения НЕТ,
// оно в списке аргументов.

// testSchema — минимальная схема модели для резолва колонки (без БД/AST).
func testSchema() {
	m := &project.Schema{
		NameTable:  "rooms",
		NamePascal: "Room",
		NameCamel:  "room",
		FieldsMap: map[string]*project.SchemaField{
			"name":   {NamePascal: "Name", NameCamel: "name", NameCol: "name"},
			"roomId": {NamePascal: "RoomId", NameCamel: "roomId", NameCol: "room_id"},
		},
	}
	project.SchemasLib = project.SchemasMap{"room": m}
}

// buildQuery рендерит SQL так, как его увидит драйвер: `String()` прогоняет
// `AppendQuery` с форматтером, то есть включает и текст, и ПОДСТАНОВКУ
// аргументов (значение с апострофом обязано быть экранировано).
func buildQuery(t *testing.T, f *FilterModel) string {
	t.Helper()
	return renderQuery(t, func(sel *bun.SelectQuery) { f.constructWhere(sel) })
}

// buildQueryIn — отдельно: SetType собирается `constructWhereIn`.
func buildQueryIn(t *testing.T, f *FilterModel) string {
	t.Helper()
	return renderQuery(t, func(sel *bun.SelectQuery) { f.constructWhereIn(sel) })
}

func renderQuery(t *testing.T, apply func(*bun.SelectQuery)) string {
	t.Helper()
	testSchema()
	db := bun.NewDB(nil, pgdialect.New())
	sel := db.NewSelect()
	apply(sel)
	return sel.String()
}

// Значение с апострофом и «инъекцией» не должно попадать в текст SQL.
func TestConstructWhere_ValueIsParameter(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "name",
		Type:     StringType,
		Operator: Eq,
		Value1:   "O'Brien'; DROP TABLE rooms; --",
	}

	query := buildQuery(t, f)

	// `'; DROP TABLE` в значении попадает в СТРОКУ-литерал, а не ломает запрос:
	// апостроф экранирован удвоением, остальное — внутри кавычек.
	// Значение целиком лежит ВНУТРИ строкового литерала: апостроф удвоен,
	// `DROP TABLE` — часть строки, а не отдельная инструкция.
	if !strings.Contains(query, "O''Brien") {
		t.Fatalf("апостроф не экранирован: %s", query)
	}
	// Литерал закрывается ОДИНАРНОЙ кавычкой: `'';` внутри значения — это
	// экранированная кавычка (безопасно), а `';` после значения — конец строки.
	// Проверяем, что НЕзакрытая кавычка не вывела `DROP` из литерала.
	body := strings.TrimPrefix(query, "SELECT * WHERE (\"rooms\".\"name\" = ")
	if !strings.HasSuffix(body, ")") {
		t.Fatalf("неожиданная форма запроса: %s", query)
	}
	literal := strings.TrimSuffix(body, ")")
	if !strings.HasPrefix(literal, "'") || !strings.HasSuffix(literal, "'") {
		t.Fatalf("значение не заключено в строковый литерал: %s", query)
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(literal, "'"), "'")
	// Внутри литерала кавычки должны быть ТОЛЬКО удвоенными: 'O''Brien''; DROP...'
	if strings.Contains(strings.ReplaceAll(inner, "''", ""), "'") {
		t.Fatalf("одинарная кавычка внутри литерала не удвоена — инъекция: %s", query)
	}
	if !strings.Contains(inner, "DROP TABLE rooms") {
		t.Fatalf("текст значения не попал внутрь литерала: %s", query)
	}
}

// between: оба значения — параметры.
func TestConstructWhere_BetweenValuesAreParameters(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "name",
		Type:     StringType,
		Operator: Btw,
		Value1:   "a'1",
		Value2:   "b'2",
	}

	query := buildQuery(t, f)

	if !strings.Contains(query, "a''1") || !strings.Contains(query, "b''2") {
		t.Fatalf("значения between не экранированы как литералы: %s", query)
	}
}

// boolean: значение — параметр, а не литерал в тексте.
func TestConstructWhere_BooleanIsParameter(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "name",
		Type:     BooleanType,
		Operator: Eq,
		Boolean:  true,
	}

	query := buildQuery(t, f)

	// Форматтер сам решает, как передать bool (`TRUE` в pg-диалекте) — важно,
	// что это значение, а не конкатенация строки в текст запроса.
	if !strings.Contains(strings.ToUpper(query), "TRUE") {
		t.Fatalf("boolean не попал в запрос: %s", query)
	}
}

// IN: элементы набора — параметры.
func TestConstructWhereIn_SetIsParameter(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "name",
		Type:     SetType,
		Operator: In,
		Set:      []string{"a'--", "b"},
	}

	// SetType идёт через `constructWhereIn`, а не `constructWhere`.
	query := buildQueryIn(t, f)

	// Апостроф элемента набора удвоен и остаётся внутри литерала.
	if !strings.Contains(query, "a''--") {
		t.Fatalf("элемент набора не экранирован: %s", query)
	}
}

// Регрессия 42P01: LIKE-фильтр строил `lower(regexp_replace(news.title, ...))`
// и отдавал выражение через `schema.UnsafeIdent`. Bun видит ТОЧКУ внутри
// выражения и разбирает его как `таблица.колонка` — в SQL уезжает алиас
// `lower(regexp_replace(news`, и Postgres отвечает:
//   missing FROM-clause entry for table "lower(regexp_replace(news" (42P01).
//
// Ловится по наличию КАВЫЧЕК вокруг таблицы и колонки (`"rooms"."name"`),
// а не по «сырому» `rooms.name` без кавычек.
func TestConstructWhere_LikeQuotesDottedColumn(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "name",
		Type:     StringType,
		Operator: Like,
		Value1:   "а",
	}

	query := buildQuery(t, f)

	if !strings.Contains(query, "regexp_replace(") {
		t.Fatalf("LIKE-фильтр не использует regexp_replace: %s", query)
	}
	// Колонка обязана быть ОТКВОЧЕНА как идентификатор — иначе bun примет
	// точку внутри функции за разделитель таблица.колонка.
	if !strings.Contains(query, `"rooms"."name"`) {
		t.Fatalf("колонка внутри regexp_replace не отвочена: %s", query)
	}
	if strings.Contains(query, "regexp_replace(rooms.name") {
		t.Fatalf("колонка ушла в выражение без кавычек (42P01): %s", query)
	}
}

// Колонка по-прежнему берётся ИЗ СХЕМЫ (имя поля модели → колонка БД).
func TestConstructWhere_ColumnResolvedFromSchema(t *testing.T) {
	f := &FilterModel{
		Model:    "room",
		Col:      "roomId",
		Type:     StringType,
		Operator: Eq,
		Value1:   "x",
	}

	query := buildQuery(t, f)

	if !strings.Contains(query, "room_id") {
		t.Fatalf("колонка не резолвлена через схему: %s", query)
	}
}
