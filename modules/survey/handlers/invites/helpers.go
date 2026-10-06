package invites

import (
	"strings"

	"github.com/pro-assistance-dev/sprob/modules/survey/notify"
)

// render — шаблон письма (общий движок с уведомлениями).
func render(template string, values map[string]string) string {
	return notify.Render(template, values)
}

// looksLikeEmail — минимальная проверка адреса (без полной RFC-валидации:
// мусор отсекаем, но экзотику не ломаем). SMTP всё равно отклонит неверный.
func looksLikeEmail(email string) bool {
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return false
	}
	domain := email[at+1:]
	return strings.Contains(domain, ".") && !strings.ContainsAny(email, " \t\n")
}

// lowerAll — список email в нижнем регистре (для сравнения без учёта регистра).
func lowerAll(emails []string) []string {
	res := make([]string, 0, len(emails))
	for _, e := range emails {
		res = append(res, strings.ToLower(strings.TrimSpace(e)))
	}
	return res
}
