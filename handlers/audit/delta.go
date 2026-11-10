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
	"id":               "ID",
	"surname":          "Фамилия",
	"name":             "Имя",
	"patronymic":       "Отчество",
	"date_birth":       "Дата рождения",
	"is_male":          "Пол",
	"code":             "Код",
	"full_code":        "Полный код",
	"comment":          "Комментарий",
	"archived":         "Архивный",
	"candidate":        "Кандидат",
	"is_student":       "Студент",
	"show":             "Показывать",
	"public_map":       "Публикация на карте",
	"public_portal":    "Публикация на портале",
	"accepted":         "Документ принят",
	"file_info_id":     "Файл",
	"worker_id":        "Сотрудник",
	"human_id":         "Человек",
	"department_id":    "Отдел",
	"position_id":      "Должность",
	"document_type_id": "Тип документа",
	"user_id":          "Пользователь",
	"photo_id":         "Фото",
	"contact_id":       "Контакт",
	"patr_name":        "Отчество",
	"first_name":       "Имя",
	"last_name":        "Фамилия",
	"middle_name":      "Отчество",
	"number":           "Номер",
	"value":            "Значение",
	"title":            "Название",
	"description":      "Описание",
	"type":             "Тип",
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
		b, _ := json.Marshal(val)
		return string(b)
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", val))
		if s == "" || s == "0001-01-01T00:00:00Z" {
			return "—"
		}
		// UUID ссылки — показываем коротко, но узнаваемо.
		if len(s) == 36 && strings.Count(s, "-") == 4 && strings.HasSuffix(field, "_id") {
			return "ссылка " + s[:8]
		}
		// Даты из БД приходят с временем — оставляем только дату.
		if len(s) >= 19 && s[4] == '-' && s[10] == 'T' {
			return s[:10] + " " + s[11:16]
		}
		return s
	}
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
	newMap = flatten(newMap)

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

// flatten — поднимает вложенные объекты на верхний уровень (кроме `id`).
//
// ⚠️ `new_value` — это ТЕЛО ЗАПРОСА, а не строка БД: в нём приходят поля в
// camelCase (`dateBirth`, `editNameMode`, `isMale`) и служебные флаги экрана.
// Сравнивать их с колонками БД (`date_birth`, `is_male`) нельзя — получались
// ложные «изменения» вида «EditNameMode: — → нет» (08.11). Имена приводим к
// змеиному виду, а «флаги UI» отбрасываем.
func flatten(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	put := func(k string, v interface{}) {
		col := toSnake(k)
		if ignoredRequestFields[col] {
			return
		}
		out[col] = v
	}
	for k, v := range m {
		if nested, ok := v.(map[string]interface{}); ok {
			for nk, nv := range nested {
				if nk == "id" {
					continue
				}
				put(nk, nv)
			}
			continue
		}
		put(k, v)
	}
	return out
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
