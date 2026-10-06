package public

import (
	"errors"
	"testing"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"
)

// resolveParams — чистая логика параметризации анкеты: обязательность и
// допустимые значения. Тест фиксирует поведение при пустых/некорректных данных.
func TestResolveParams(t *testing.T) {
	params := models.SurveyParams{
		{Key: "department", Label: "Отделение", FieldCode: "dept", Required: true, Options: []string{"cardio", "surgery"}},
		{Key: "source", Label: "Источник"},
	}

	t.Run("valid value resolves and prefills", func(t *testing.T) {
		query := map[string]string{"department": "cardio", "source": "email"}
		resolved, prefill, err := resolveParams(params, query)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved["department"] != "cardio" || resolved["source"] != "email" {
			t.Fatalf("resolved = %v", resolved)
		}
		if prefill["dept"] != "cardio" {
			t.Fatalf("prefill = %v", prefill)
		}
	})

	t.Run("missing required param fails with typed error (400, not 500)", func(t *testing.T) {
		_, _, err := resolveParams(params, map[string]string{})
		if err == nil {
			t.Fatal("expected error for missing required param")
		}
		// ⚠️ Регрессия: раньше это был обычный fmt.Errorf → handler отдавал
		// 500. Теперь ошибка типизированная и маппится в 400.
		if !errors.Is(err, ErrParamRequired) {
			t.Fatalf("expected ErrParamRequired, got %v", err)
		}
	})

	t.Run("value outside options fails with typed error", func(t *testing.T) {
		query := map[string]string{"department": "unknown"}
		_, _, err := resolveParams(params, query)
		if !errors.Is(err, ErrParamInvalid) {
			t.Fatalf("expected ErrParamInvalid, got %v", err)
		}
	})

	t.Run("optional missing param is skipped", func(t *testing.T) {
		query := map[string]string{"department": "surgery"}
		resolved, _, err := resolveParams(params, query)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := resolved["source"]; ok {
			t.Fatalf("optional param should be absent, got %v", resolved)
		}
	})
}
