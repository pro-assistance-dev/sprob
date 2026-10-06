package notify

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Операторы условий совпадают с FieldRule (sprob) — единая семантика «если, то».
const (
	opEquals      = "equals"
	opNotEquals   = "notEquals"
	opContains    = "contains"
	opNotContains = "notContains"
	opAnswered    = "answered"
	opNotAnswered = "notAnswered"
	opGreater     = "greater"
	opLess        = "less"
)

// RulesPass — выполнены ли ВСЕ условия уведомления. Пусто — всегда true.
func RulesPass(rules models.NotificationRules, values map[string]string) bool {
	for _, rule := range rules {
		if !rulePass(rule, values) {
			return false
		}
	}
	return true
}

// fillRawValue — «сырое» значение ответа для условий: id варианта (radio),
// id вариантов через "|" (set), иначе строка/число (совпадает с `NotificationRule.Value`).
func fillRawValue(fill *forms.FieldFill) string {
	if len(fill.SelectedAnswerVariants) > 0 {
		parts := make([]string, 0, len(fill.SelectedAnswerVariants))
		for _, selected := range fill.SelectedAnswerVariants {
			if selected.AnswerVariantID.Valid {
				parts = append(parts, selected.AnswerVariantID.UUID.String())
			}
		}
		return strings.Join(parts, "|")
	}
	if fill.AnswerVariantId.Valid {
		return fill.AnswerVariantId.UUID.String()
	}
	if fill.ValueString != "" {
		return fill.ValueString
	}
	if fill.ValueNumber != 0 {
		return strconv.FormatFloat(float64(fill.ValueNumber), 'f', -1, 32)
	}
	return ""
}

func rulePass(rule *models.NotificationRule, values map[string]string) bool {
	value := values["field:"+rule.FieldCode]
	answered := value != ""

	// ⚠️ Операторы сравнения с пустым `Value` не настраиваемы осмысленно:
	// без этой защиты `equals` с пустым ожиданием СОВПАДАЛ с неотвеченным
	// вопросом ('' == '') — уведомление уходило на каждый ответ. Пустое значение
	// в условии сравнения считаем невыполненным.
	comparesValue := rule.Operator == opEquals || rule.Operator == opNotEquals ||
		rule.Operator == opContains || rule.Operator == opNotContains ||
		rule.Operator == opGreater || rule.Operator == opLess
	if comparesValue && rule.Value == "" {
		return false
	}

	switch rule.Operator {
	case opAnswered:
		return answered
	case opNotAnswered:
		return !answered
	case opEquals:
		return answered && value == rule.Value
	case opNotEquals:
		return answered && value != rule.Value
	case opContains:
		return answered && contains(strings.Split(value, "; "), rule.Value)
	case opNotContains:
		return answered && !contains(strings.Split(value, "; "), rule.Value)
	case opGreater:
		return answered && toFloat(value) > toFloat(rule.Value)
	case opLess:
		return answered && toFloat(value) < toFloat(rule.Value)
	default:
		return true
	}
}

// Render — подстановка {{ключ}} из values (шаблоны писем вне Notifier).
func Render(template string, values map[string]string) string {
	return render(template, values)
}

// render подставляет {{ключ}} из values. Неизвестный ключ оставляем как есть,
// чтобы опечатка была видна в письме/логе, а не молча исчезала.
func render(template string, values map[string]string) string {
	return placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}"), " ")
		if value, ok := values[key]; ok {
			return value
		}
		return match
	})
}

var placeholderRe = regexp.MustCompile(`\{\{\s*[a-zA-Z0-9_:.-]+\s*\}\}`)

// formFields — плоский список вопросов формы (по секциям, в порядке).
func (n *Notifier) formFields(ctx context.Context, publicationID uuid.NullUUID) (forms.Fields, error) {
	publication := models.Publication{}
	if err := n.db.NewSelect().Model(&publication).Where("id = ?", publicationID).Scan(ctx); err != nil {
		return nil, err
	}
	form := forms.Form{}
	if err := n.db.NewSelect().
		Model(&form).
		Relation("FormSections", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("form_sections.item_order")
		}).
		Relation("FormSections.Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("fields.item_order")
		}).
		Relation("FormSections.Fields.AnswerVariants").
		Where("forms.id = ?", publication.FormID).
		Scan(ctx); err != nil {
		return nil, err
	}
	fields := make(forms.Fields, 0)
	for _, section := range form.FormSections {
		fields = append(fields, section.Fields...)
	}
	return fields, nil
}

// variantNames — карта id варианта → название.
func (n *Notifier) variantNames(fields forms.Fields) map[uuid.NullUUID]string {
	names := make(map[uuid.NullUUID]string)
	for _, field := range fields {
		for _, variant := range field.AnswerVariants {
			names[variant.ID] = variant.Name
		}
	}
	return names
}

// fillText — текстовое представление ответа (radio/set — названия, иначе строка).
func fillText(fill *forms.FieldFill, variantNames map[uuid.NullUUID]string) string {
	if fill.AnswerVariantId.Valid {
		if name, ok := variantNames[fill.AnswerVariantId]; ok {
			return name
		}
	}
	if len(fill.SelectedAnswerVariants) > 0 {
		parts := make([]string, 0, len(fill.SelectedAnswerVariants))
		for _, selected := range fill.SelectedAnswerVariants {
			if name, ok := variantNames[selected.AnswerVariantID]; ok {
				parts = append(parts, name)
			}
		}
		return strings.Join(parts, "; ")
	}
	if fill.ValueString != "" {
		return fill.ValueString
	}
	if fill.ValueNumber != 0 {
		return strconv.FormatFloat(float64(fill.ValueNumber), 'f', -1, 32)
	}
	return ""
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func toFloat(value string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return f
}

func parseUUID(id string) uuid.NullUUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: parsed, Valid: true}
}
