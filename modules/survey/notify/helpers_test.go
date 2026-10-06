package notify

import (
	"testing"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

func rule(fieldCode, operator, value string) *models.NotificationRule {
	return &models.NotificationRule{FieldCode: fieldCode, Operator: operator, Value: value}
}

func TestRulesPass(t *testing.T) {
	values := map[string]string{"field:department": "Кардиология", "field:score": "4"}

	// Пусто — всегда true.
	if !RulesPass(models.NotificationRules{}, values) {
		t.Fatal("no rules must pass")
	}

	// equals: совпадает/не совпадает.
	if !RulesPass(models.NotificationRules{rule("department", "equals", "Кардиология")}, values) {
		t.Fatal("equals must pass")
	}
	if RulesPass(models.NotificationRules{rule("department", "equals", "Хирургия")}, values) {
		t.Fatal("equals mismatch must fail")
	}

	// Несколько правил — AND.
	rules := models.NotificationRules{
		rule("department", "equals", "Кардиология"),
		rule("score", "greater", "3"),
	}
	if !RulesPass(rules, values) {
		t.Fatal("AND of true rules must pass")
	}
	rules = append(rules, rule("score", "less", "1"))
	if RulesPass(rules, values) {
		t.Fatal("AND with a false rule must fail")
	}
}

// Пустое значение условия сравнения НЕ выполняется, даже если вопрос не отвечен
// (`equals ""` не должен давать ложное срабатывание).
func TestRulesPassEmptyValue(t *testing.T) {
	empty := map[string]string{}
	for _, op := range []string{"equals", "notEquals", "contains", "notContains", "greater", "less"} {
		if RulesPass(models.NotificationRules{rule("department", op, "")}, empty) {
			t.Fatalf("rule %q with empty value must not pass", op)
		}
	}
	// А вот equals с НЕПУСТЫМ значением и неотвеченным вопросом — false.
	if RulesPass(models.NotificationRules{rule("department", "equals", "X")}, empty) {
		t.Fatal("equals non-empty on unanswered must fail")
	}
	// answered/notAnswered пустого значения не требуют.
	if !RulesPass(models.NotificationRules{rule("department", "notAnswered", "")}, empty) {
		t.Fatal("notAnswered on unanswered must pass")
	}
}

func TestRender(t *testing.T) {
	values := map[string]string{"field:department": "Кардиология", "param:cabinet": "12"}
	got := render("Отделение: {{field:department}}, каб. {{ param:cabinet }}", values)
	want := "Отделение: Кардиология, каб. 12"
	if got != want {
		t.Fatalf("render = %q, want %q", got, want)
	}

	// Неизвестный ключ остаётся как есть (видно опечатку).
	if got := render("{{nope}}", values); got != "{{nope}}" {
		t.Fatalf("unknown placeholder must stay, got %q", got)
	}
}
