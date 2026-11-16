package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DeltaItem — ОДНО изменённое поле в человекочитаемом виде.
//
// ⚠️ ЗАЧЕМ (08.11, замечание: «нужно человекочитаемое было/стало — именно
// дельта»). До этого в журнале лежали два JSON-снимка всего объекта
// (`old_value`/`new_value`), и по записи было не понять, ЧТО поменялось:
// приходилось глазами сравнивать два длинных JSON.
type DeltaItem struct {
	Field string `json:"field"` // сырое имя колонки
	Label string `json:"label"` // «Фамилия»
	Old   string `json:"old"`   // «Выхристюк»
	New   string `json:"new"`   // «Иванова»
}

// fieldLabels — русские подписи колонок.
//
// ⚠️ Держим рядом с рендером, а не в БД: журнал читают люди, и подпись — часть
// представления. Нет в словаре — покажем имя колонки с `_`→` ` и с большой
// буквы (не идеально, но читаемо), а не пустоту.
var fieldLabels = map[string]string{
	"id":            "ID",
	"surname":       "Фамилия",
	"name":          "Имя",
	"patronymic":    "Отчество",
	"date_birth":    "Дата рождения",
	"is_male":       "Пол",
	"code":          "Код",
	"full_code":     "Полный код",
	"comment":       "Комментарий",
	"archived":      "Архивный",
	"candidate":     "Кандидат",
	"is_student":    "Студент",
	"show":          "Показывать",
	"public_map":    "Публикация на карте",
	"public_portal": "Публикация на портале",
	"accepted":      "Документ принят",

	"worker_id":     "Сотрудник",
	"human_id":      "Человек",
	"department_id": "Отдел",
	"position_id":   "Должность",

	"user_id":     "Пользователь",
	"photo_id":    "Фото",
	"contact_id":  "Контакт",
	"patr_name":   "Отчество",
	"first_name":  "Имя",
	"last_name":   "Фамилия",
	"middle_name": "Отчество",
	"number":      "Номер",
	"value":       "Значение",
	// ⚠️ Поля документов сотрудника (10.11): в журнале они выглядели как
	// сырые `Fio`, `Date`, `Number` — английские ключи словаря документов.
	"fio":           "ФИО в документе",
	"date":          "Дата в документе",
	"issued":        "Кем выдан",
	"issued_date":   "Когда выдан",
	"division_code": "Код подразделения",
	"passport_code": "Код подразделения",
	"field_id":      "Поле документа",

	"document_id":      "Документ",
	"item_order":       "Порядок",
	"mask_id":          "Маска",
	"is_required":      "Обязательное",
	"file_system_path": "Путь к файлу",

	"title":       "Название",
	"description": "Описание",
	"type":        "Тип",
}

// humanValue — как показать значение поля.
//
// ⚠️ Спец-случаи, без которых журнал «врёт»:
//   - `null`/пусто → «—» (а не пустая строка: тогда «было → стало» выглядит
//     как «ничего → ничего»);
//   - UUID внешней связи → «ссылка <id8>» (пользователю не нужен полный id,
//     но нужно отличать «сменили связь» от «очистили»);
//   - bool → «да/нет»;
//   - `file_info_id` и прочие id → «ссылка», чтобы не читать сырой uuid.
func humanValue(field string, v interface{}) string {
	if v == nil {
		return "—"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "да"
		}
		return "нет"
	case float64:
		// JSON-числа: 1 → «1», а не «1.000000».
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%v", val)
	case []interface{}:
		if len(val) == 0 {
			return "—"
		}
		parts := make([]string, 0, len(val))
		for _, it := range val {
			parts = append(parts, humanValue(field, it))
		}
		return strings.Join(parts, ", ")
	case map[string]interface{}:
		// ⚠️ Вложенный объект в дельте — это НЕ значение поля, а «связанная
		// сущность» (в теле запроса клиент присылает `worker: {...}`,
		// `contact: {...}`, `documentType: {...}` целиком). Печатать их JSON
		// в «стало» бессмысленно: получалось
		// «Contact: — → {"description":"","emails":[…]}» — владелец
		// справедливо не понял, что изменилось (15.11).
		// Показываем ЧЕЛОВЕЧЕСКУЮ подпись связи, если её удаётся собрать
		// (ФИО/название/код).
		if label := nestedLabel(val); label != "" {
			return label
		}
		// ⚠️ Подписи нет, но объект — это ссылка с `id`: показываем единым
		// видом «ссылка <8 симв.>», как у снимка БД. Иначе одна и та же
		// связь («Сотрудник: ссылка a717770d» → «ссылка») считалась
		// изменённой (замер 15.11).
		if id := strOf(val["id"]); isUUID(id) {
			return "ссылка " + id[:8]
		}
		return "ссылка"
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", val))
		if s == "" || s == "0001-01-01T00:00:00Z" {
			return "—"
		}
		// UUID — это ССЫЛКА на другую сущность. Показываем единым видом
		// («ссылка 227614fd») независимо от того, пришла ли она из снимка БД
		// (`document_type_id`) или из тела запроса (`fileInfo: {id}`).
		// ⚠️ Раньше без этого правила одна и та же связь сравнивалась как
		// «ссылка 227614fd → ссылка» и попадала в дельту ложным изменением
		// (замер 15.11). Если подпись удаётся собрать (`nestedLabel`), до
		// этого места дело не доходит — там уже ФИО/название.
		if isUUID(s) {
			return "ссылка " + s[:8]
		}
		// Даты из БД приходят с временем — оставляем только дату.
		if len(s) >= 19 && s[4] == '-' && s[10] == 'T' {
			return s[:10] + " " + s[11:16]
		}
		return s
	}
}

// nestedLabel — человекочитаемая подпись СВЯЗАННОГО объекта из тела запроса.
//
// ⚠️ Вложенные объекты в теле — это связи (`worker`, `contact`, `position`,
// `documentType`), и печатать их JSON'ом бессмысленно (см. `humanValue`).
// Собираем подпись по известным полям в порядке «самое узнаваемое»:
// ФИО (в `human`), название, код. Ничего не нашли — пустая строка, и
// вызывающий покажет «ссылка».
func nestedLabel(v map[string]interface{}) string {
	// ФИО живёт внутри `human` (или плоско в самом объекте).
	for _, holder := range []map[string]interface{}{v, asMap(v["human"])} {
		if holder == nil {
			continue
		}
		name := strings.TrimSpace(strings.Join(nonEmpty(
			strOf(holder["surname"]), strOf(holder["name"]), strOf(holder["patronymic"]),
		), " "))
		if name != "" {
			return name
		}
		if full := strOf(holder["fullName"]); full != "" {
			return full
		}
	}
	for _, key := range []string{"name", "title", "label", "code"} {
		if s := strOf(v[key]); s != "" {
			return s
		}
	}
	return ""
}

// isUUID — строка выглядит как UUID (8-4-4-4-12).
func isUUID(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

func asMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func strOf(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// deltaForOperation — какую дельту показывает запись журнала.
//
// ⚠️ ЗАЧЕМ (15.11). Владелец: «если было создание сущности — не нужно
// показывать создание каждого из полей». До этого дельта строилась ОДИНАКОВО
// для всех операций, и у создания (снимка «до» нет) каждое поле выглядело как
// «— → значение»: «Создание · Id — → a1b2…, Contact — → {…}, CreatedAt — → …».
// Это не «что изменилось», а «вот все поля объекта».
//
// Правила:
//   - create   — ПУСТО: имя объекта в колонке «Что сделал» уже отвечает на
//                вопрос «что произошло»;
//   - update   — дельта полей (единственное, где она содержательна);
//   - delete   — что БЫЛО в записи (`old` → пусто), `id` отсекает `BuildDelta`;
//   - откат    — пусто: служебная запись, а не правка полей.
//
// ⚠️ Пустой результат — `[]`, а не `nil`: фронт делает `.length` без проверок.
func deltaForOperation(operation, oldValue, newValue string, isRevert bool) []DeltaItem {
	if isRevert || operation == "create" {
		return []DeltaItem{}
	}
	if operation == "delete" {
		return buildDeleted(oldValue)
	}
	return BuildDelta(oldValue, newValue)
}

// buildDeleted — что было в удалённой записи.
//
// ⚠️ Через `BuildDelta(old, "")` не получается (15.11): там поля есть в «до»,
// но отсутствуют в «после», и `BuildDelta` считает их «не трогали» — на
// удалении это ВСЕ поля, и дельта выходила пустой. Здесь наоборот: пустое
// «после» — это и есть факт удаления, поэтому показываем заполненные поля
// снимка «до».
// ⚠️ `id` и служебные технические поля не показываем: «Идентификатор: было
// удалено» не отвечает на «что удалили».
func buildDeleted(oldRaw string) []DeltaItem {
	oldMap := parseJSONMap(oldRaw)
	out := make([]DeltaItem, 0, len(oldMap))
	for k, v := range oldMap {
		if k == "id" || isTechnicalField(k) {
			continue
		}
		val := humanValue(k, v)
		if val == "" || val == "—" {
			continue
		}
		out = append(out, DeltaItem{Field: k, Label: fieldLabel(k), Old: val, New: "—"})
	}
	sort.Slice(out, func(i, j int) bool {
		pi, pj := deltaPriority(out[i].Field), deltaPriority(out[j].Field)
		if pi != pj {
			return pi < pj
		}
		return out[i].Label < out[j].Label
	})
	return out
}

// isTechnicalField — служебные поля, которые не описывают суть записи.
//
// ⚠️ Держим рядом с правилом удаления: «Что удалили» должно отвечать на
// вопрос про ОБЪЕКТ, а `created_at`/`updated_by` — про учёт.
func isTechnicalField(col string) bool {
	switch col {
	case "created_at", "updated_at", "created_by", "updated_by", "search", "is_new":
		return true
	}
	return strings.HasSuffix(col, "_id") || strings.HasSuffix(col, "Id")
}

// BuildDelta — дельта между снимками «до» и «после».
//
// ⚠️ Сравниваем по КОЛОНКАМ, а не текстом JSON: порядок ключей в JSON не
// гарантирован, и текстовое сравнение давало ложные «изменения» на каждой
// записи.
func BuildDelta(oldRaw, newRaw string) []DeltaItem {
	oldMap := parseJSONMap(oldRaw)
	newMap := parseJSONMap(newRaw)

	// `new_value` у update-записей может быть телом запроса, где поля вложены
	// (`human: {...}`, `workerPosition: {...}`) — раскладываем их на верхний
	// уровень, иначе сравнение с плоским снимком даёт мусор.
	oldKeys := make(map[string]bool, len(oldMap))
	for k := range oldMap {
		oldKeys[k] = true
	}
	newMap = flatten(newMap, oldKeys)

	keys := make(map[string]bool, len(oldMap)+len(newMap))
	for k := range oldMap {
		// ⚠️ Поле есть в снимке «до», но НЕТ в теле запроса — его НЕ трогали.
		// `new_value` — это то, что клиент отправил, а не полный объект: без
		// этой проверки в дельте появлялся мусор «Пол: нет → —» на каждой
		// правке фамилии (08.11).
		if _, present := newMap[k]; !present {
			continue
		}
		keys[k] = true
	}
	for k := range newMap {
		keys[k] = true
	}

	// ⚠️ Пустой результат — `nil`, и JSON отдаёт `null`: фронт делает `.length`
	// и `.map` без проверок, поэтому «нет изменений» обязано быть `[]`.
	out := make([]DeltaItem, 0, len(keys))
	for k := range keys {
		if k == "id" {
			continue // id не меняется, в дельту не идёт
		}
		oldV := humanValue(k, oldMap[k])
		newV := humanValue(k, newMap[k])
		if oldV == newV {
			continue
		}
		out = append(out, DeltaItem{Field: k, Label: fieldLabel(k), Old: oldV, New: newV})
	}
	// Стабильный порядок: сначала «человек» (ФИО), потом остальное по алфавиту.
	sort.Slice(out, func(i, j int) bool {
		pi, pj := deltaPriority(out[i].Field), deltaPriority(out[j].Field)
		if pi != pj {
			return pi < pj
		}
		return out[i].Label < out[j].Label
	})
	return out
}

// deltaPriority — важные для человека поля идут первыми.
func deltaPriority(field string) int {
	switch field {
	case "surname", "last_name":
		return 0
	case "name", "first_name":
		return 1
	case "patronymic", "patr_name", "middle_name":
		return 2
	}
	return 10
}

func fieldLabel(col string) string {
	if l, ok := fieldLabels[col]; ok {
		return l
	}
	// ⚠️ ПУСТАЯ колонка — паника `col[:1]` (index out of range). Фильтр может
	// прийти без `col` (или с пустым `field` у старой записи журнала), и это
	// валило ВСЮ ручку чтения журнала 500-й (поймано 10.11 после переноса в sprob).
	if col == "" {
		return ""
	}
	return strings.ToUpper(col[:1]) + strings.ReplaceAll(col[1:], "_", " ")
}

func parseJSONMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

// flatten — приводит тело запроса к виду, сравнимому со снимком БД.
//
// ⚠️ ЗАЧЕМ ГЛУБОКАЯ ПЕРЕРАБОТКА (15.11). Разбор живого журнала hr показал, что
// дельта была НЕЧИТАЕМОЙ по двум причинам сразу:
//
//  1. `old_value` — ПЛОСКИЙ снимок строки (`document_type_id`, `worker_id`),
//     а `new_value` — ПОЛНОЕ тело от клиента (`documentType: {…}`,
//     `worker: {…, human: {…}}`). Названия полей разные (`document_type_id` vs
//     `documentTypeId`), и сравнение давало мусор.
//  2. Прежний `flatten` ПОДНИМАЛ вложенные объекты на верхний уровень — из
//     `worker: {human: {surname}}` в дельту попадали `Surname: — → …`,
//     `Archived: — → нет` и ещё 20 строк на ОДНУ правку галочки.
//
// Новая логика:
//   - вложенные объекты НЕ разворачиваем: это СВЯЗЬ, её подпись даёт
//     `humanValue` → `nestedLabel` («Прививка от кори», ФИО);
//   - id-ссылки приводим к паре «имя поля + _id» (`documentTypeId` →
//     `documentType` со значением-ссылкой), чтобы совпасть со снимком;
//   - поля, которых НЕ БЫЛО в снимке «до», в дельту не идут: клиент присылает
//     весь объект, но у документа нет колонки `archived` — это поле
//     сотрудника из вложенного `worker`.
func flatten(m map[string]interface{}, oldKeys map[string]bool) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		col := toSnake(k)
		if ignoredRequestFields[col] {
			continue
		}
		if nested, ok := v.(map[string]interface{}); ok {
			// ⚠️ Связь, а не набор полей: кладём объект целиком под её именем
			// (значение свернёт `humanValue` → `nestedLabel`).
			// ⚠️ Имя связи приводим к виду КОЛОНКИ из снимка: клиент шлёт
			// `documentType`, а в БД `document_type_id`. Без общего ключа одна
			// и та же связь сравнивалась сама с собой («ссылка» у обеих сторон
			// или ложное «— → ссылка»).
			key := alignKey(col, oldKeys)
			// ⚠️ Связи, которой НЕТ в снимке этой сущности, в дельту не берём.
			// Тело документа сотрудника содержит `worker: {…}` целиком, но
			// колонки `worker` у `worker_documents` нет — правку отметки
			// приёма «Сотрудник: — → ссылка» не объясняло (замер 15.11).
			if len(oldKeys) > 0 && !oldKeys[key] && !oldKeys[col+"_id"] {
				continue
			}
			out[key] = nested
			continue
		}
		// ⚠️ Поля, которого НЕТ в снимке «до», в дельту не идёт: клиент шлёт
		// весь объект, включая вычисляемые поля UI (`docsAccepted`,
		// `phonesCount`) и поля вложенных сущностей. Без этого отсева одна
		// правка галочки давала 23 строки «— → нет» (замер 15.11).
		// ⚠️ Исключение — `create` (снимка нет вовсе): там `oldKeys` пуст, и
		// отсев не применяется.
		key := alignKey(col, oldKeys)
		if len(oldKeys) > 0 && !oldKeys[key] {
			continue
		}
		out[key] = v
	}
	return out
}

// alignKey — привести имя поля из тела запроса к имени КОЛОНКИ снимка.
//
// ⚠️ Расхождение форм (15.11): клиент шлёт `documentTypeId`, `workerId`,
// `employmentFormId`, а снимок БД — `document_type_id`, `worker_id`,
// `employment_form_id`. Без сопоставления одна и та же связь считалась
// изменённой («— → ссылка»), а настоящее изменение маскировалось.
//
// Варианты в порядке вероятности: `col` как есть, `col + "_id"` (связь),
// версия с `_uuid` (легаси). Первое совпадение со снимком и берём.
func alignKey(col string, oldKeys map[string]bool) string {
	if len(oldKeys) == 0 || oldKeys[col] {
		return col
	}
	for _, candidate := range []string{col + "_id", col + "_uuid"} {
		if oldKeys[candidate] {
			return candidate
		}
	}
	return col
}

// ignoredRequestFields — поля тела, которым НЕ соответствует колонка БД:
// служебные флаги форм (`editNameMode`), вычисляемые поля и метаданные.
var ignoredRequestFields = map[string]bool{
	"edit_name_mode":    true,
	"edit_contact_mode": true,
	"is_new":            true,
	"created_at":        true,
	"updated_at":        true,
	"created_by":        true,
	"updated_by":        true,
	"search":            true,
}

// toSnake — `dateBirth` → `date_birth`; уже змеиный оставляем как есть.
func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
