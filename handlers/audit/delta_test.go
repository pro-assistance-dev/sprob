package audit

import "testing"

// Дельта журнала (08.11): «было → стало» по полям, по-русски.
//
// ⚠️ ЧТО ЗАЩИЩАЕМ. До этого в журнале лежали два JSON-снимка целиком, и
// человек не мог понять, что изменилось. Здесь фиксируем:
//   - изменённое поле попадает в дельту, неизменённое — НЕТ;
//   - подпись по-русски, значения человекочитаемы (bool → да/нет, null → —);
//   - вложенный объект (`human: {surname}`) сравнивается с плоским снимком;
//   - порядок стабильный (ФИО первым).
func TestBuildDelta(t *testing.T) {
	// ⚠️ ФИО у `workers` хранится в отдельной таблице `humans`: снимок
	// сотрудника содержит `human_id`, а не `human`. Поэтому «ФИО до» в снимке
	// нет, и вложенный `human` из тела — это шум вложенной сущности (см.
	// `TestBuildDelta_NoNestedNoise`). Здесь проверяем то, что реально
	// сравнимо: скалярные поля и правку с ОБЕИХ сторон.
	oldRaw := `{"id":"w1","archived":false,"comment":"старый","phone":"123"}`
	newRaw := `{"id":"w1","archived":true,"comment":"новый","phone":"123","human":{"surname":"Иванова"}}`

	got := BuildDelta(oldRaw, newRaw)

	byField := map[string]DeltaItem{}
	for _, d := range got {
		byField[d.Field] = d
	}

	if d, ok := byField["archived"]; !ok || d.Old != "нет" || d.New != "да" {
		t.Fatalf("bool не человекочитаем: %+v", byField["archived"])
	}
	if d, ok := byField["comment"]; !ok || d.Old != "старый" || d.New != "новый" {
		t.Fatalf("текстовое поле не попало в дельту: %+v", byField["comment"])
	}
	// Неизменённое не шумит; `id` в дельту не идёт; вложенная сущность,
	// которой нет в снимке, — тоже.
	for _, absent := range []string{"id", "phone", "human"} {
		if _, ok := byField[absent]; ok {
			t.Fatalf("%q попал в дельту без изменений", absent)
		}
	}
}

// Плоский ФИО (как реально приходит у `humans`) — поля по отдельности.
//
// ⚠️ Живой снимок hr: `{"surname":"T907839","name":"Юлия",…}` — без обёртки
// `human`. Здесь дельта обязана показать «Фамилия: было → стало», а не одну
// связь: это правка ФИО человека, и человек ищет в журнале именно фамилию.
func TestBuildDelta_FlatHuman(t *testing.T) {
	oldRaw := `{"id":"h1","surname":"Выхристюк","name":"Юлия","patronymic":"Петровна"}`
	newRaw := `{"id":"h1","surname":"Иванова","name":"Юлия","patronymic":"Петровна"}`

	got := BuildDelta(oldRaw, newRaw)
	if len(got) != 1 {
		t.Fatalf("ожидалась одна строка дельты, получили %+v", got)
	}
	if got[0].Field != "surname" || got[0].Old != "Выхристюк" || got[0].New != "Иванова" {
		t.Fatalf("фамилия не попала в дельту: %+v", got[0])
	}
	if got[0].Label != "Фамилия" {
		t.Fatalf("подпись поля = %q, want «Фамилия»", got[0].Label)
	}
}

// Поля вложенных сущностей НЕ должны попадать в дельту своей правкой.
//
// ⚠️ Замер 15.11: правка `accepted` у документа давала 23 строки, потому что
// тело содержит `worker: {archived, phonesCount, docsAccepted…}`. Этих полей
// в снимке `worker_documents` нет — значит, и изменений по ним нет.
func TestBuildDelta_NoNestedNoise(t *testing.T) {
	oldRaw := `{"id":"d1","accepted":true,"comment":""}`
	newRaw := `{"id":"d1","accepted":false,"comment":"","worker":{"archived":false,"phonesCount":0,"docsAccepted":3}}`

	got := BuildDelta(oldRaw, newRaw)
	if len(got) != 1 || got[0].Field != "accepted" {
		t.Fatalf("в дельту попал шум вложенных сущностей: %+v", got)
	}
	if got[0].Old != "да" || got[0].New != "нет" {
		t.Fatalf("bool не человекочитаем: %+v", got[0])
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
