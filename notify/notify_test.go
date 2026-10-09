package notify

import (
	"os"
	"path/filepath"
	"strings"
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
	data := map[string]any{"entity": "order", "action": "created", "status": "new"}
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

func TestRenderRichData(t *testing.T) {
	event := Event{
		Entity:  "orders",
		Action:  "created",
		Payload: map[string]any{"status": "new"},
		Data:    map[string]any{"Number": "42"},
	}
	d := renderData(event)
	if d["status"] != "new" {
		t.Errorf("status = %v", d["status"])
	}
	if got := execTemplate("Заказ {{.data.Number}}", d); got != "Заказ 42" {
		t.Errorf("data-шаблон = %q", got)
	}
}

func TestRenderFileTemplate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TEMPLATES_PATH", dir)
	if err := os.WriteFile(filepath.Join(dir, "_header.html"), []byte("H|"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "body.gohtml"), []byte(`{{template "_header.html" .}}{{.status}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := renderFile("body.gohtml", map[string]any{"status": "new"})
	if err != nil {
		t.Fatalf("renderFile: %v", err)
	}
	if !strings.Contains(out, "H|") || !strings.Contains(out, "new") {
		t.Errorf("renderFile = %q", out)
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
