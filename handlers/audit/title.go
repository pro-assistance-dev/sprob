package audit

import (
	"context"
	"fmt"
	"strings"
)

// titleQueries — как получить ЧЕЛОВЕКОЧИТАЕМОЕ имя объекта по сущности журнала.
//
// ⚠️ ЗАЧЕМ (10.11, замечание владельца: «id для обычного юзера ничего не
// говорит»). В журнале был только `objectId` (uuid) — по нему нельзя понять,
// ЧТО именно изменили. Здесь — имя для каждой сущности: ФИО сотрудника,
// название отдела, тип документа и т. д.
//
// ⚠️ Почему запрос на сущность, а не «одно поле в таблице»: у разных таблиц
// имя строится по-разному (ФИО из трёх колонок, название — из одной). Правило
// описывается в SQL-выражении, поэтому остаётся локальным и не требует правок
// в моделях.
var titleQueries = map[string]struct {
	table string
	// expr — готовое SQL-выражение имени; пусто — собирается из колонок
	// `title`/`name` по фактической схеме (см. `exprFor`).
	expr string
}{
	"humans": {table: "humans",
		expr: `nullif(trim(concat_ws(' ', surname, name, patronymic)), '')`},
	"workers": {table: "workers w",
		expr: `(SELECT nullif(trim(concat_ws(' ', h.surname, h.name, h.patronymic)), '')
		        FROM humans h WHERE h.id = w.human_id)`},
	"departments":            {table: "departments", expr: "name"},
	"positions":              {table: "positions", expr: "name"},
	"document-types":         {table: "document_types", expr: "name"},
	"documents":              {table: "documents", expr: "name"},
	"worker-documents":       {table: "worker_documents", expr: "code"},
	"worker-document-fields": {table: "worker_document_fields", expr: "value"},
	"worker-positions":       {table: "worker_positions", expr: "id::text"},
	"users":                  {table: "users", expr: "name"},
	// ⚠️ У портала имя — `title`, у rdkb — `name`. `expr` НЕ задаём: выражение
	// собирается по фактическим колонкам в `exprFor` (иначе SQL падал на
	// несуществующей колонке, и `objectTitle` оставался пустым — 11.11).
	"news":         {table: "news"},
	"pages":        {table: "pages"},
	"publications": {table: "publications"},
	"rooms":        {table: "rooms", expr: "name"},
	"buildings":    {table: "buildings", expr: "name"},
	"floors":       {table: "floors", expr: "name"},
	"polygons":     {table: "polygons", expr: "name"},
	"faq":          {table: "faq", expr: "question"},
}

// titleCache — кэш имён в пределах ОДНОГО запроса (в списке журнала одни и те
// же объекты встречаются десятки раз, и дёргать БД на каждую строку wasteful).
type titleKey struct{ entity, id string }

// objectTitle — имя объекта для строки журнала.
//
// ⚠️ Если имя получить нельзя (нет таблицы/колонки, строка удалена — что
// бывает после `revert` создания), возвращаем:
//   - `code` записи журнала, если он есть («пакет: 3 объекта», код документа);
//   - иначе пустую строку — фронт покажет название сущности без имени.
//
// ⚠️ Никогда не показываем голый uuid: это и была претензия.
func (a *Audit) objectTitle(ctx context.Context, it *Item) string {
	if it.ObjectID == "" {
		return firstNonEmpty(it.Code, "")
	}
	spec, ok := titleQueries[it.Entity]
	if !ok {
		return firstNonEmpty(it.Code, "")
	}

	key := titleKey{it.Entity, it.ObjectID}
	if v, ok := a.titles[key]; ok {
		return v
	}

	expr, err := a.exprFor(ctx, spec)
	if err != nil || expr == "" {
		a.titles[key] = firstNonEmpty(it.Code, "")
		return a.titles[key]
	}

	var rows []struct {
		Title *string `bun:"title"`
	}
	q := fmt.Sprintf(`SELECT %s AS title FROM %s WHERE id = ?`, expr, spec.table)
	if err := a.helper.DB.IDB(ctx).NewRaw(q, it.ObjectID).Scan(ctx, &rows); err != nil || len(rows) == 0 {
		a.titles[key] = firstNonEmpty(it.Code, "")
		return a.titles[key]
	}
	title := firstNonEmpty(deref(rows[0].Title), it.Code, "")
	a.titles[key] = title
	return title
}

// exprFor — SQL-выражение имени объекта.
//
// ⚠️ Если у сущности задано `expr` — берём его (ФИО, специальные случаи).
// Иначе собираем по ФАКТИЧЕСКИМ колонкам таблицы: `title` (портал) и/или
// `name` (rdkb). Так список не падает, если колонки нет — и не требует
// отдельной карты на каждый проект.
func (a *Audit) exprFor(ctx context.Context, spec struct {
	table string
	expr  string
}) (string, error) {
	if spec.expr != "" {
		return spec.expr, nil
	}
	cols, err := a.scalarColumns(ctx, spec.table)
	if err != nil {
		return "", err
	}
	has := func(c string) bool {
		for _, x := range cols {
			if x == c {
				return true
			}
		}
		return false
	}
	switch {
	case has("title") && has("name"):
		return `coalesce(nullif(title, ''), nullif(name, ''))`, nil
	case has("title"):
		return `nullif(title, '')`, nil
	case has("name"):
		return `nullif(name, '')`, nil
	}
	return "", nil
}

// resetTitles — очистить кэш имён перед формированием страницы.
func (a *Audit) resetTitles() {
	a.titles = make(map[titleKey]string)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
