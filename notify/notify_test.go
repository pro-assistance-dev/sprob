package notify

import (
	"testing"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

func TestEventMatch(t *testing.T) {
	ev := Event{Entity: "order", Action: models.ActionCreated}
	cases := []struct {
		pattern string
		want    bool
	}{
		{"", true},
		{"*", true},
		{"*.*", true},
		{"order.created", true},
		{"order.*", true},
		{"*.created", true},
		{"order.updated", false},
		{"event.created", false},
		{"event.*", false},
	}
	for _, c := range cases {
		if got := ev.Match(c.pattern); got != c.want {
			t.Errorf("Match(%q) = %v, want %v", c.pattern, got, c.want)
		}
	}
}

func TestRulePasses(t *testing.T) {
	ev := Event{
		Entity:  "order",
		Action:  models.ActionCreated,
		Payload: map[string]any{"status": "new", "total": 5000},
	}
	// Нет условий — проходит всегда.
	if !rulePasses(&models.Rule{}, ev) {
		t.Fatal("пустое правило должно проходить")
	}
	// Условие выполнено.
	ok := &models.Rule{Filter: models.Conds{{Field: "status", Operator: "equals", Value: "new"}}}
	if !rulePasses(ok, ev) {
		t.Fatal("status=new должно проходить")
	}
	// Условие не выполнено.
	bad := &models.Rule{Filter: models.Conds{{Field: "status", Operator: "equals", Value: "done"}}}
	if rulePasses(bad, ev) {
		t.Fatal("status=done не должно проходить")
	}
	// AND: одно из двух не выполнено → не проходит.
	and := &models.Rule{Filter: models.Conds{
		{Field: "status", Operator: "equals", Value: "new"},
		{Field: "total", Operator: "contains", Value: "999"},
	}}
	if rulePasses(and, ev) {
		t.Fatal("AND с непройденным условием не должно проходить")
	}
}

func TestExecTemplate(t *testing.T) {
	data := map[string]string{"entity": "order", "action": "created", "status": "new"}
	if got := execTemplate("{{.entity}}.{{.action}}", data); got != "order.created" {
		t.Errorf("шаблон = %q", got)
	}
	if got := execTemplate("Статус: {{.status}}", data); got != "Статус: new" {
		t.Errorf("шаблон = %q", got)
	}
	// Битый шаблон не роняет отправку — отдаём как есть.
	if got := execTemplate("{{.broken", data); got != "{{.broken" {
		t.Errorf("битый шаблон = %q", got)
	}
}

func TestBackoff(t *testing.T) {
	if backoff(1) != 1*60*1e9 {
		t.Errorf("backoff(1) != 1m")
	}
	if backoff(4) != 8*60*1e9 {
		t.Errorf("backoff(4) != 8m")
	}
}
