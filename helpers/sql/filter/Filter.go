package filter

import (
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

func (items FilterModels) mergeJoins() {
	joinModels := make(map[string]*FilterModel)
	for i := range items {
		if items[i].Type == JoinType && items[i].Operator == In {
			items[i].JoinIn = []string{items[i].Col}
			items[i].JoinSets = [][]string{items[i].Set}

			findedModel, ok := joinModels[items[i].JoinTable]

			if ok {

				findedModel.JoinIn = append(findedModel.JoinIn, items[i].Col)
				findedModel.JoinSets = append(findedModel.JoinSets, items[i].Set)

				items[i].ignore = true
			} else {
				joinModels[items[i].JoinTable] = items[i]
			}

		}
	}
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
			filterModel.constructJoin(query)
		default:
			return
		}
	}

	// OR-группы применяем в порядке появления — результат детерминирован.
	// `WhereGroup("(", fn)`: условия внутри fn объединяются через `Where`/`WhereOr`.
	for _, g := range groupOrder {
		frags := groups[g]
		query = query.WhereGroup("(", func(q *bun.SelectQuery) *bun.SelectQuery {
			for i, fr := range frags {
				if i > 0 {
					q = q.WhereOr(fr.Query, fr.Args...)
					continue
				}
				q = q.Where(fr.Query, fr.Args...)
			}
			return q
		})
	}
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
