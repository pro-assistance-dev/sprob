package notify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

func TestSign(t *testing.T) {
	payload := []byte(`{"a":1}`)
	got := sign(payload, "s3cr3t")
	// Проверяем руками.
	mac := hmac.New(sha256.New, []byte("s3cr3t"))
	mac.Write(payload)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("sign = %q, want %q", got, want)
	}
	// Разный секрет → разная подпись.
	if sign(payload, "other") == got {
		t.Fatal("подпись не должна совпадать при другом секрете")
	}
}

func TestWebhookRulesPass(t *testing.T) {
	values := map[string]string{"field:dept": "cardio"}
	rules := models.WebhookRules{
		{FieldCode: "dept", Operator: "equals", Value: "cardio"},
	}
	if !webhookRulesPass(rules, values) {
		t.Fatal("правило должно выполняться")
	}
	rules = append(rules, &models.WebhookRule{FieldCode: "dept", Operator: "equals", Value: "surgery"})
	if webhookRulesPass(rules, values) {
		t.Fatal("AND со вторым (ложным) правилом должен давать false")
	}
	// Пусто правил — вызов на каждый ответ.
	if !webhookRulesPass(models.WebhookRules{}, values) {
		t.Fatal("без правил — всегда true")
	}
}
