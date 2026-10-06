package reports

import (
	"context"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
)

// Service — аналитика и выгрузка ответов по публикации.
type Service struct {
	r *Repository
}

// GetPublicationByForm возвращает публикацию, привязанную к форме (или nil, если нет).
// Нужно для таба «Публикация»: страница формы находит свой бланк по formId.
func (s *Service) GetPublicationByForm(c context.Context, formID string) (*models.Publication, error) {
	return s.r.GetPublicationByForm(c, uuid.NullUUID{UUID: uuid.MustParse(formID), Valid: true})
}

// GetResponsesByPublication возвращает ответы публикации (таб «Ответы»).
// filters — значения параметров анкеты для отбора (key → value).
func (s *Service) GetResponsesByPublication(c context.Context, publicationID string, filters map[string]string) (models.Responses, error) {
	return s.r.GetResponses(c, uuid.NullUUID{UUID: uuid.MustParse(publicationID), Valid: true}, filters)
}

// Summary — сводка по публикации: счётчик ответов, разрез по параметрам,
// агрегаты по каждому вопросу (распределение вариантов / список значений).
type Summary struct {
	Publication *models.Publication `json:"publication"`
	Total       int                 `json:"total"`
	// ByParams — разрез по параметрам анкеты: key → value → число ответов.
	ByParams map[string]map[string]int `json:"byParams"`
	// Fields — агрегаты по вопросам в порядке формы.
	Fields []FieldSummary `json:"fields"`
}

// FieldSummary — агрегат по одному вопросу.
type FieldSummary struct {
	FieldID    uuid.NullUUID  `json:"fieldId"`
	Name       string         `json:"name"`
	Code       string         `json:"code"`
	ValueType  string         `json:"valueType"`
	Answered   int            `json:"answered"`
	Counts     map[string]int `json:"counts,omitempty"`
	NumberAvg  float64        `json:"numberAvg,omitempty"`
	NumberSum  float64        `json:"numberSum,omitempty"`
	TextValues []string       `json:"textValues,omitempty"`
}

// GetSummary собирает сводку по публикации.
func (s *Service) GetSummary(c context.Context, publicationID string) (*Summary, error) {
	publication, err := s.r.GetPublication(c, publicationID)
	if err != nil {
		return nil, err
	}
	form, err := s.r.GetForm(c, publication.FormID)
	if err != nil {
		return nil, err
	}
	responses, err := s.r.GetResponses(c, publication.ID, nil)
	if err != nil {
		return nil, err
	}

	fields := formFields(form)
	variantNames := collectVariantNames(fields)

	summary := &Summary{
		Publication: publication,
		Total:       len(responses),
		ByParams:    map[string]map[string]int{},
		Fields:      make([]FieldSummary, 0, len(fields)),
	}

	// Разрез по параметрам, объявленным у публикации.
	paramKeys := make([]string, 0, len(publication.Params))
	for _, param := range publication.Params {
		paramKeys = append(paramKeys, param.Key)
		summary.ByParams[param.Key] = map[string]int{}
	}

	counts := make([]map[string]int, len(fields))
	numbers := make([]float64, len(fields))
	answered := make([]int, len(fields))
	texts := make([][]string, len(fields))

	for _, response := range responses {
		for _, key := range paramKeys {
			if value := response.Params[key]; value != "" {
				summary.ByParams[key][value]++
			}
		}
		fills := response.FormFill.FieldFills
		for index, field := range fields {
			for _, fill := range fills {
				if fill.FieldID != field.ID {
					continue
				}
				if counts[index] == nil {
					counts[index] = map[string]int{}
				}
				answered[index]++
				collectFieldValue(field, fill, variantNames, counts[index], &numbers[index], &texts[index])
				break
			}
		}
	}

	for index, field := range fields {
		valueType := ""
		if field.ValueType != nil {
			valueType = string(field.ValueType.Name)
		}
		item := FieldSummary{
			FieldID:   field.ID,
			Name:      field.Name,
			Code:      field.Code,
			ValueType: valueType,
			Answered:  answered[index],
			Counts:    counts[index],
		}
		if field.ValueType != nil && field.ValueType.IsNumber() {
			item.NumberSum = numbers[index]
			if answered[index] > 0 {
				item.NumberAvg = numbers[index] / float64(answered[index])
			}
		}
		item.TextValues = texts[index]
		summary.Fields = append(summary.Fields, item)
	}

	return summary, nil
}

// ExportXlsx формирует книгу Excel: лист «Ответы» (строка = ответ, колонки =
// параметры + вопросы) и лист «Сводка» (агрегаты по вопросам).
func (s *Service) ExportXlsx(c context.Context, publicationID string) ([]byte, string, error) {
	publication, err := s.r.GetPublication(c, publicationID)
	if err != nil {
		return nil, "", err
	}
	form, err := s.r.GetForm(c, publication.FormID)
	if err != nil {
		return nil, "", err
	}
	responses, err := s.r.GetResponses(c, publication.ID, nil)
	if err != nil {
		return nil, "", err
	}

	fields := formFields(form)
	variantNames := collectVariantNames(fields)

	header := []string{"Дата"}
	for _, param := range publication.Params {
		header = append(header, labelOrKey(param))
	}
	for _, field := range fields {
		header = append(header, field.Name)
	}

	rows := make([][]interface{}, 0, len(responses))
	for _, response := range responses {
		row := make([]interface{}, 0, len(header))
		row = append(row, response.CreatedAt.Format("2006-01-02 15:04"))
		for _, param := range publication.Params {
			row = append(row, response.Params[param.Key])
		}
		for _, field := range fields {
			row = append(row, fieldAnswerText(field, response.FormFill.FieldFills, variantNames))
		}
		rows = append(rows, row)
	}

	summary, err := s.GetSummary(c, publicationID)
	if err != nil {
		return nil, "", err
	}
	summaryRows := make([][]interface{}, 0, len(summary.Fields))
	for _, field := range summary.Fields {
		detail := ""
		if len(field.Counts) > 0 {
			detail = formatCounts(field.Counts)
		} else if len(field.TextValues) > 0 {
			detail = formatCounts(countValues(field.TextValues))
		} else if field.Answered > 0 {
			detail = "среднее: " + formatFloat(field.NumberAvg)
		}
		summaryRows = append(summaryRows, []interface{}{field.Name, field.Answered, detail})
	}

	sheets := []writersSheet{
		{Name: "Ответы", Header: header, Rows: rows},
		{Name: "Сводка", Header: []string{"Вопрос", "Ответов", "Распределение"}, Rows: summaryRows},
	}

	data, err := buildWorkbook(sheets)
	if err != nil {
		return nil, "", err
	}
	return data, fileName(publication), nil
}

// fieldAnswerText — текстовое представление ответа на вопрос (для выгрузки).
func fieldAnswerText(field *forms.Field, fills forms.FieldFills, variantNames map[uuid.NullUUID]string) string {
	for _, fill := range fills {
		if fill.FieldID != field.ID {
			continue
		}
		return fillText(field, fill, variantNames)
	}
	return ""
}

// collectFieldValue добавляет вклад ответа в агрегаты вопроса.
func collectFieldValue(field *forms.Field, fill *forms.FieldFill, variantNames map[uuid.NullUUID]string, counts map[string]int, numberSum *float64, textValues *[]string) {
	if field.ValueType == nil {
		return
	}
	switch {
	case field.ValueType.IsRadio():
		if name, ok := variantNames[fill.AnswerVariantId]; ok {
			counts[name]++
		}
	case field.ValueType.IsSet():
		for _, selected := range fill.SelectedAnswerVariants {
			if name, ok := variantNames[selected.AnswerVariantID]; ok {
				counts[name]++
			}
		}
	case field.ValueType.IsNumber():
		*numberSum += float64(fill.ValueNumber)
	default:
		if fill.ValueString != "" {
			*textValues = append(*textValues, fill.ValueString)
		}
	}
}
