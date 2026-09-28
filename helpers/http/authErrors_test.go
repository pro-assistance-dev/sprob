package http

import (
	"errors"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

// Классификация ошибок токена (pros: «неверный/истёкший токен = 500, а не 401»).
//
// ⚠️ Реальные сообщения из логов прода — именно они должны давать 401:
//
//	Error #01: Token is expired
//	Error #01: signature is invalid
func TestIsAuthError_JWTValidation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			// Точное совпадение работало и раньше — регрессия недопустима.
			"просроченный (точная строка)",
			errors.New("Token is expired"),
			true,
		},
		{
			// ⚠️ ГЛАВНЫЙ баг: jwt-go добавляет «by <duration>» — точное
			// сравнение не срабатывало и клиент получал 500.
			"просроченный (с длительностью)",
			errors.New("Token is expired by 1h2m3s"),
			true,
		},
		{
			"невалидная подпись",
			errors.New("signature is invalid"),
			true,
		},
		{
			"неизвестный алгоритм",
			errors.New("unexpected signing method: RS256"),
			true,
		},
		{
			"битый токен",
			errors.New("token contains an invalid number of segments"),
			true,
		},
		{
			"токен ещё не действителен",
			errors.New("Token is not valid yet"),
			true,
		},
		{
			"нет claim",
			errors.New("claim not found"),
			true,
		},
		{
			// Типизированная ошибка jwt-go (самый надёжный путь)
			"ValidationError(expired)",
			&jwt.ValidationError{Errors: jwt.ValidationErrorExpired},
			true,
		},
		{
			"ValidationError(signature)",
			&jwt.ValidationError{Errors: jwt.ValidationErrorSignatureInvalid},
			true,
		},
		{
			// Обёрнутая ошибка (jwt нередко заворачивают) — errors.As должен пробить.
			"обёрнутый ValidationError",
			wrapErr(&jwt.ValidationError{Errors: jwt.ValidationErrorMalformed}, "parse"),
			true,
		},

		// ── НЕ auth-ошибки: должны остаться 500 ──────────────────────────────
		{"обычная ошибка БД", errors.New("connection refused"), false},
		{"nil", nil, false},
		{"ошибка валидации тела", errors.New("invalid character 'x'"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestStatusForError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"просроченный токен → 401", errors.New("Token is expired by 1h"), http.StatusUnauthorized},
		{"подпись → 401", errors.New("signature is invalid"), http.StatusUnauthorized},
		{"сбой БД → 500", errors.New("dial tcp: connection refused"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusForError(tt.err); got != tt.want {
				t.Errorf("StatusForError(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

// wrapErr имитирует обёртку (fmt.Errorf с %w) вокруг ошибки jwt.
func wrapErr(err error, msg string) error {
	return &wrappedError{inner: err, msg: msg}
}

type wrappedError struct {
	inner error
	msg   string
}

func (w *wrappedError) Error() string { return w.msg + ": " + w.inner.Error() }
func (w *wrappedError) Unwrap() error { return w.inner }
