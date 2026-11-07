package filter

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"strings"
)

func (items FilterModels) mergeJoins() {
	// ⚠️ 🔴 НЕ заполняем `JoinIn`/`JoinSets` из `Col`: `constructJoinV3` уже
	// строит условие по `Col` (и только когда `Set` непуст), а старый путь
	// добавлял ВТОРОЕ такое же `JoinOn` — дубль `IN` в ON и «отдел ≠ должность»
	// при двух фильтрах одной таблицы (жалоба 08.11: фильтры «не работают»).
	//
	// Сейчас функция ничего не склеивает: второй join-фильтр по той же таблице
	// остаётся отдельным элементом и добавляет условие в ОДИН `JOIN ON`, а сам
	// `JOIN` повторно не делается (`FilterModel.Joined`, см. `CreateFilter`).
	_ = items
}

func (items FilterModels) CreateFilter(query *bun.SelectQuery) {
	if len(items) == 0 {
		return
	}

	items.mergeJoins()

	// Фрагменты OR-групп собираем ОТДЕЛЬНО: фильтры одной группы идут в один
	// `WhereGroup` с OR, остальные — обычным `Where` (AND).
	groups := make(map[int][]schema.QueryWithArgs)
	var groupOrder []int

	// Join-таблицы, уже присоединённые в ЭТОМ запросе: повторный `JOIN` той же
	// таблицы Postgres отвергает («table name specified more than once»), а
	// условия второго фильтра добавляются в тот же `JOIN ON`.
	joined := make(map[string]bool)

	for _, filterModel := range items {
		if filterModel.ignore {
			continue
		}

		switch filterModel.Type {
		case SetType:
			if len(filterModel.Set) == 0 {
				break
			}
			// ⚠️ Set/In в OR-группу пока не кладём (нужна отдельная обработка
			// `IN` внутри группы) — игнорируем, а не молча ломаем SQL.
			if filterModel.Group != 0 {
				break
			}
			filterModel.constructWhereIn(query)
		case DateType:
			filterModel.datesToString()
			items.apply(filterModel, query, groups, &groupOrder)
		case StringType, BooleanType, NumberType:
			items.apply(filterModel, query, groups, &groupOrder)
		case JoinType:
			// Метку ставит `CreateFilter`, а не `constructJoinV3`: модель
			// переиспользуется между запросами (сортировка прогоняет запрос
			// дважды), и флаг на ней «залипал» бы — второй запрос потерял бы JOIN.
			if filterModel.JoinTableModel == "" {
				continue
			}
			if joined[filterModel.JoinTableModel] {
				filterModel.Joined = true
				filterModel.constructJoin(query)
				continue
			}
			joined[filterModel.JoinTableModel] = true
			filterModel.Joined = false
			filterModel.constructJoin(query)
		default:
			return
		}
	}

	// OR-группы применяем в порядке появления — результат детерминирован.
	//
	// ⚠️ Собираем группу ОДНИМ фрагментом `(A OR B)`, а не `query.WhereGroup`:
	// `WhereGroup` после уже добавленного условия клеит группу БЕЗ `AND` —
	// `WHERE (A)((B OR C))` → Postgres 42601 «syntax error at or near (»
	// (07.11: поиск «иванов» вместе с «без email» давал 500). Обычный
	// `query.Where("(…)", args…)` сам ставит `AND` между условиями.
	for _, g := range groupOrder {
		frags := groups[g]
		if len(frags) == 0 {
			continue
		}
		frag := orGroupFragment(frags)
		query.Where(frag.Query, frag.Args...)
	}
}

// orGroupFragment склеивает фрагменты одной OR-группы в `(A OR B OR …)`.
// Значения — параметры (`?`), поэтому склейка текста безопасна.
func orGroupFragment(frags []schema.QueryWithArgs) schema.QueryWithArgs {
	var b strings.Builder
	args := make([]interface{}, 0, len(frags))
	b.WriteString("(")
	for i, fr := range frags {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString("(")
		b.WriteString(fr.Query)
		b.WriteString(")")
		args = append(args, fr.Args...)
	}
	b.WriteString(")")
	return schema.QueryWithArgs{Query: b.String(), Args: args}
}

// apply применяет фрагмент фильтра: в OR-группу или сразу в запрос (AND).
func (items FilterModels) apply(f *FilterModel, query *bun.SelectQuery, groups map[int][]schema.QueryWithArgs, order *[]int) {
	if f.Group == 0 {
		f.constructWhere(query)
		return
	}
	frag, ok := f.whereFragment()
	if !ok {
		return
	}
	if _, seen := groups[f.Group]; !seen {
		*order = append(*order, f.Group)
	}
	groups[f.Group] = append(groups[f.Group], frag)
}
