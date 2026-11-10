package audit

import "testing"

// Дельта журнала (08.11): «было → стало» по полям, по-русски.
//
// ⚠️ ЧТО ЗАЩИЩАЕМ. До этого в журнале лежали два JSON-снимка целиком, и
// человек не мог понять, что изменилось. Здесь фиксируем:
//   • изменённое поле попадает в дельту, неизменённое — НЕТ;
//   • подпись по-русски, значения человекочитаемы (bool → да/нет, null → —);
//   • вложенный объект (`human: {surname}`) сравнивается с плоским снимком;
//   • порядок стабильный (ФИО первым).
func TestBuildDelta(t *testing.T) {
	oldRaw := `{"id":"w1","surname":"Выхристюк","name":"Юлия","archived":false,"comment":null}`
	newRaw := `{"human":{"surname":"Иванова","name":"Юлия","patronymic":"Петровна"},"archived":true,"comment":null,"id":"w1"}`

	got := BuildDelta(oldRaw, newRaw)

	byField := map[string]DeltaItem{}
	for _, d := range got {
		byField[d.Field] = d
	}

	if d, ok := byField["surname"]; !ok || d.Old != "Выхристюк" || d.New != "Иванова" {
		t.Fatalf("фамилия не попала в дельту: %+v", byField["surname"])
	}
	if byField["surname"].Label != "Фамилия" {
		t.Fatalf("подпись поля = %q, want «Фамилия»", byField["surname"].Label)
	}
	if d, ok := byField["archived"]; !ok || d.Old != "нет" || d.New != "да" {
		t.Fatalf("bool не человекочитаем: %+v", byField["archived"])
	}
	if d, ok := byField["patronymic"]; !ok || d.Old != "—" || d.New != "Петровна" {
		t.Fatalf("новое поле не показано как «—» → значение: %+v", byField["patronymic"])
	}
	// Неизменённое не шумит; `id` в дельту не идёт.
	if _, ok := byField["name"]; ok {
		t.Fatal("неизменённое поле попало в дельту")
	}
	if _, ok := byField["id"]; ok {
		t.Fatal("id попал в дельту")
	}
	// ФИО — первым.
	if len(got) == 0 || got[0].Field != "surname" {
		t.Fatalf("первым должно идти ФИО, получили %+v", got)
	}
}

// Записи БЕЗ снимка «до» (всё, что старше 08.11) не должны давать ложную дельту.
func TestBuildDelta_NoOldSnapshot(t *testing.T) {
	got := BuildDelta("", `{"surname":"Иванова"}`)
	for _, d := range got {
		if d.Old != "—" {
			t.Fatalf("без снимка «до» старое должно быть «—»: %+v", d)
		}
	}
}

// Ссылки на внешние сущности показываем коротко, но узнаваемо.
func TestHumanValue_Reference(t *testing.T) {
	got := humanValue("department_id", "e3f61a88-da63-433a-840f-fe837dd92733")
	if got != "ссылка e3f61a88" {
		t.Fatalf("ссылка = %q, want «ссылка e3f61a88»", got)
	}
}

// Пустая колонка не должна ронять подпись (10.11: паника `col[:1]` давала 500
// на всей ручке чтения журнала).
func TestFieldLabel_Empty(t *testing.T) {
	if got := fieldLabel(""); got != "" {
		t.Fatalf("fieldLabel(\"\") = %q, want пусто", got)
	}
	if got := fieldLabel("date_birth"); got != "Дата рождения" {
		t.Fatalf("fieldLabel(date_birth) = %q", got)
	}
}
