package filter

import (
	"github.com/uptrace/bun"
)

// CreateFilter func
func (i *Filter) CreateFilter(query *bun.SelectQuery) {
	if len(i.FilterModels) == 0 {
		return
	}
	for _, filterModel := range i.FilterModels {
		switch filterModel.Type {
		case SetType:
			if len(filterModel.Set) == 0 {
				break
			}
			filterModel.constructWhereIn(query)
		case DateType:
			filterModel.datesToString()
			filterModel.constructWhere(query)
		case StringType, BooleanType, NumberType:
			filterModel.constructWhere(query)
		case JoinType:
			filterModel.constructJoin(query)
		// case "number":
		//	tbl = constructNumberWhere(tbl, field, filter)
		// case "text":
		//	if filterOperator == "" {
		//		tbl = constructTextWhere(tbl, field, filterOperator, filter)
		//	} else {
		//		tbl = constructTextWhere(tbl, field, filterOperator, filter.Condition1.filter, filter.Condition2.filter)
		//	}
		default:
			// log.Println("unknown number filterType: " + *filter.FilterType)
			return
		}
	}
}
func (items FilterModels) mergeJoins() {
	joinModels := make(map[string]int)
	for i := range items {
		if items[i].Type == JoinType && items[i].Operator == In {
			index, ok := joinModels[items[i].Model]
			if ok {
				items[index].Set = append(items[index].Set, items[i].Set...)
				items[i].ignore = true
			} else {
				joinModels[items[i].Model] = i
			}
		}
	}
}
func (items FilterModels) CreateFilter(query *bun.SelectQuery) {
	if len(items) == 0 {
		return
	}

	items.mergeJoins()

	for _, filterModel := range items {
		if filterModel.ignore {
			continue
		}

		switch filterModel.Type {
		case SetType:
			if len(filterModel.Set) == 0 {
				break
			}
			filterModel.constructWhereIn(query)
		case DateType:
			filterModel.datesToString()
			filterModel.constructWhere(query)
		case StringType, BooleanType, NumberType:
			filterModel.constructWhere(query)
		case JoinType:
			filterModel.constructJoin(query)
		// case "number":
		//	tbl = constructNumberWhere(tbl, field, filter)
		// case "text":
		//	if filterOperator == "" {
		//		tbl = constructTextWhere(tbl, field, filterOperator, filter)
		//	} else {
		//		tbl = constructTextWhere(tbl, field, filterOperator, filter.Condition1.filter, filter.Condition2.filter)
		//	}
		default:
			// log.Println("unknown number filterType: " + *filter.FilterType)
			return
		}
	}
}
