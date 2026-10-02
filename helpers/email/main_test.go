package email

import (
	"strings"
	"testing"

	"github.com/pro-assistance-dev/sprob/config"
)

func newTestEmail() *Email {
	e := NewEmail(config.Email{From: "pro-assistance@example.org", Server: "smtp.example.org", Port: "465"})
	e.request = request{To: []string{"user@example.com"}, Subject: "Приглашаем на мероприятие «Тест»", Body: "<p>Привет</p>"}
	return e
}

// Заголовки, без которых массовая рассылка уходит в спам, обязаны быть в письме.
func TestBuildMessageHasAntiSpamHeaders(t *testing.T) {
	msg, err := newTestEmail().buildMessage()
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	for _, h := range []string{"Date:", "Message-ID:", "List-Unsubscribe:", "List-Unsubscribe-Post:", "MIME-Version: 1.0"} {
		if !strings.Contains(msg, h) {
			t.Errorf("нет заголовка %q:\n%s", h, msg)
		}
	}
}

// Кириллица в теме обязана быть RFC 2047-кодирована, а не сырой.
func TestBuildMessageEncodesCyrillicSubject(t *testing.T) {
	msg, err := newTestEmail().buildMessage()
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	if !strings.Contains(msg, "Subject: =?utf-8?") {
		t.Fatalf("тема не закодирована:\n%s", msg)
	}
	if strings.Contains(msg, "Subject: Приглашаем") {
		t.Fatalf("тема ушла сырой кириллицей:\n%s", msg)
	}
}

// Заголовки должны идти в детерминированном порядке (From первым).
func TestBuildMessageHeaderOrderStable(t *testing.T) {
	first, err := newTestEmail().buildMessage()
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	if !strings.HasPrefix(first, "From: ") {
		t.Fatalf("From не первый заголовок:\n%s", first)
	}
	for i := 0; i < 20; i++ {
		next, err := newTestEmail().buildMessage()
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		got := messageIDIn(next)
		firstID := messageIDIn(first)
		if got == "" || firstID == "" {
			t.Fatal("Message-ID отсутствует")
		}
	}
}

func messageIDIn(msg string) string {
	for _, line := range strings.Split(msg, "\r\n") {
		if strings.HasPrefix(line, "Message-ID: ") {
			return strings.TrimPrefix(line, "Message-ID: ")
		}
	}
	return ""
}
