package qr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pro-assistance-dev/sprob/modules/links/models"
)

// Тесты чистых функций домена: короткий код и резолвер целей.
// Они не требуют БД — проверяют контракт, на который опираются map/food.

func TestNewCode_Format(t *testing.T) {
	seen := map[string]bool{}
	for range 1000 {
		code, err := NewCode()
		if err != nil {
			t.Fatalf("NewCode: %v", err)
		}
		if len(code) != codeLength {
			t.Fatalf("длина кода %d, ожидалось %d (%q)", len(code), codeLength, code)
		}
		for _, ch := range code {
			if !containsChar(alphabet, ch) {
				t.Fatalf("символ %q вне алфавита (%q)", ch, code)
			}
		}
		if seen[code] {
			t.Fatalf("повтор кода за 1000 генераций: %q", code)
		}
		seen[code] = true
	}
}

func TestStaticResolver(t *testing.T) {
	r := StaticResolver{Rules: map[string]string{
		"map/room":   "https://map.rdkb-mos.ru/buildings?roomId={id}",
		"food/diet":  "https://food.rdkb-mos.ru/diets/{id}",
		"food/survey": "https://survey.rdkb-mos.ru/f?diet={id}",
	}}
	cases := []struct {
		name string
		link *models.Link
		want string
	}{
		{"map room", &models.Link{App: "map", TargetType: "room", TargetID: "abc"}, "https://map.rdkb-mos.ru/buildings?roomId=abc"},
		{"food diet", &models.Link{App: "food", TargetType: "diet", TargetID: "d1"}, "https://food.rdkb-mos.ru/diets/d1"},
		{"external", &models.Link{App: "links", TargetType: "external", ExternalURL: "https://ozon.ru/x"}, "https://ozon.ru/x"},
	}
	for _, tc := range cases {
		got, err := r.Resolve(t.Context(), tc.link)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != tc.want {
			t.Fatalf("%s: получили %q, ждали %q", tc.name, got, tc.want)
		}
	}
}

func TestChain_UnknownTarget(t *testing.T) {
	chain := Chain{StaticResolver{Rules: map[string]string{}}}
	if _, err := chain.Resolve(t.Context(), &models.Link{App: "map", TargetType: "unknown", TargetID: "1"}); err == nil {
		t.Fatal("неизвестная цель должна давать ошибку, а не пустой URL")
	}
}

// TestQR_ProducesPNG — дымовой тест картинки: PNG с QR кладётся в /tmp,
// распознавание проверяется скриптом (bin/linkscheck) настоящим детектором.
func TestQR_ProducesPNG(t *testing.T) {
	m, err := Encode("https://links.rdkb-mos.ru/r/abc234")
	if err != nil {
		t.Fatalf("qrEncode: %v", err)
	}
	if len(m) < 21 {
		t.Fatalf("матрица слишком мала: %d", len(m))
	}
	out := filepath.Join(os.TempDir(), "links-qr-test.png")
	if err := os.WriteFile(out, m.PNG(512), 0o644); err != nil {
		t.Fatalf("запись png: %v", err)
	}
	if svg := m.SVG(512); len(svg) < 200 {
		t.Fatalf("svg подозрительно мал: %d байт", len(svg))
	}
	t.Logf("png: %s", out)
}

func containsChar(set string, ch rune) bool {
	for _, c := range set {
		if c == ch {
			return true
		}
	}
	return false
}
