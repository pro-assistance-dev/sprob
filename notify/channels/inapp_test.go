package channels

import "testing"

func TestIsUserID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"7afec894-2376-47fd-a211-6ab59e003d17", true}, // uuid
		{"12345", true},                                // numeric sub
		{"", false},
		{"a@b.ru", false},     // email → группа/не пользователь
		{"714738997", true},   // tg chat numeric — тоже «похоже на id»
		{"admin", false},      // роль
		{"room-12", false},    // буквы кроме hex
	}
	for _, c := range cases {
		if got := isUserID(c.in); got != c.want {
			t.Errorf("isUserID(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
