package reports

import (
	"fmt"
	"sort"
	"strings"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/helpers/writers"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
)

type writersSheet = writers.XlsxSheet

func buildWorkbook(sheets []writersSheet) ([]byte, error) {
	return writers.XlsxSheets(sheets)
}

// LabelOrKey — подпись параметра для колонки (или ключ, если подписи нет).
func labelOrKey(param *models.SurveyParam) string {
	if param.Label != "" {
		return param.Label
	}
	return param.Key
}

// formFields возвращает плоский список полей формы (по секциям, в порядке).
func formFields(form *forms.Form) forms.Fields {
	fields := make(forms.Fields, 0)
	for _, section := range form.FormSections {
		fields = append(fields, section.Fields...)
	}
	return fields
}

// collectVariantNames строит карту id варианта → название (для агрегатов/выгрузки).
func collectVariantNames(fields forms.Fields) map[uuid.NullUUID]string {
	names := make(map[uuid.NullUUID]string)
	for _, field := range fields {
		for _, variant := range field.AnswerVariants {
			names[variant.ID] = variant.Name
		}
	}
	return names
}

// fillText — текстовое представление ответа на вопрос.
func fillText(field *forms.Field, fill *forms.FieldFill, variantNames map[uuid.NullUUID]string) string {
	if field.ValueType == nil {
		return ""
	}
	switch {
	case field.ValueType.IsRadio():
		return variantNames[fill.AnswerVariantId]
	case field.ValueType.IsSet():
		parts := make([]string, 0, len(fill.SelectedAnswerVariants))
		for _, selected := range fill.SelectedAnswerVariants {
			if name, ok := variantNames[selected.AnswerVariantID]; ok {
				parts = append(parts, name)
			}
		}
		return strings.Join(parts, "; ")
	case field.ValueType.IsNumber():
		return formatFloat(float64(fill.ValueNumber))
	default:
		return fill.ValueString
	}
}

// formatCounts — «Вариант: 3; Вариант: 1» в порядке убывания.
func formatCounts(counts map[string]int) string {
	type pair struct {
		key   string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for key, count := range counts {
		pairs = append(pairs, pair{key, count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, fmt.Sprintf("%s: %d", p.key, p.count))
	}
	return strings.Join(parts, "; ")
}

// countValues — частоты текстовых значений (для свободных ответов).
func countValues(values []string) map[string]int {
	counts := make(map[string]int)
	for _, value := range values {
		counts[value]++
	}
	return counts
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d", int64(value))
	}
	return fmt.Sprintf("%.2f", value)
}

// fileName — имя файла выгрузки по slug публикации.
func fileName(publication *models.Publication) string {
	slug := publication.Slug
	if slug == "" {
		slug = "survey"
	}
	return slug + "-responses.xlsx"
}
