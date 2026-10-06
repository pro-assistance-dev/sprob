package invites

import "testing"

func TestLooksLikeEmail(t *testing.T) {
	valid := []string{"a@b.ru", "user.name@example.com", "x@y.co"}
	for _, e := range valid {
		if !looksLikeEmail(e) {
			t.Errorf("%q должен считаться валидным", e)
		}
	}
	invalid := []string{"", "no-at", "@nodomain", "user@", "a@b", "a b@c.ru", "a@b .ru"}
	for _, e := range invalid {
		if looksLikeEmail(e) {
			t.Errorf("%q должен быть отброшен", e)
		}
	}
}

func TestNewTokenUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		tok := newToken()
		if len(tok) != 32 {
			t.Fatalf("токен должен быть 32 hex-символа, получили %d (%q)", len(tok), tok)
		}
		if seen[tok] {
			t.Fatalf("токен повторился: %q", tok)
		}
		seen[tok] = true
	}
}

func TestRenderInvite(t *testing.T) {
	got := render("Здравствуйте{{fio}}! {{publication}}: {{link}}", map[string]string{
		"fio":         ", Иван",
		"publication": "Опрос",
		"link":        "https://x/f/s?invite=t",
	})
	want := "Здравствуйте, Иван! Опрос: https://x/f/s?invite=t"
	if got != want {
		t.Fatalf("render = %q, want %q", got, want)
	}
}
