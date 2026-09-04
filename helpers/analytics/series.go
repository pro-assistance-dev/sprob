package analytics

// LabelValue — пара «наименование → значение» для графиков и выгрузок
// (идентична rdkb map/hr handlers/analytics; вынесена 04.09, А5.3).
type LabelValue struct {
	Label string `json:"label" bun:"label"`
	Value int    `json:"value" bun:"value"`
}

// SeriesRows превращает серию в строки плоской выгрузки с общим заголовком
// секции: {section, item.Label, item.Value}.
func SeriesRows(section string, items []LabelValue) [][]interface{} {
	rows := make([][]interface{}, 0, len(items))
	for _, it := range items {
		rows = append(rows, []interface{}{section, it.Label, it.Value})
	}
	return rows
}

// SeriesValues превращает серию в строки {label, value} без секции.
func SeriesValues(items []LabelValue) [][]interface{} {
	rows := make([][]interface{}, 0, len(items))
	for _, it := range items {
		rows = append(rows, []interface{}{it.Label, it.Value})
	}
	return rows
}
