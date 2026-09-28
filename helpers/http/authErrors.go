package http

import (
	"errors"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// IsAuthError — ошибка АУТЕНТИФИКАЦИИ (невалидный/просроченный токен), а не сбой сервера.
//
// Зачем отдельная функция. Раньше `HandleError` сравнивал строку ТОЧНО:
// `err.Error() == "Token is expired"`. Отсюда — жалоба «неверный/истёкший токен
// отдаётся как 500, а не 401»: у jwt ошибка обёрнута в `*jwt.ValidationError`
// (`Token is expired by 1h2m`, `signature is invalid`, `token is unverifiable`,
// `token contains an invalid number of segments`…), и точное сравнение не
// срабатывало ни для одного случая, кроме буквально совпадающей строки.
//
// Разбираем ДВА способа: типизированный (`errors.As` по `*jwt.ValidationError` —
// надёжно, не зависит от текста) и текстовый (страховка: свой токен-хелпер и
// сторонние обёртки отдают просто `errors.New`).
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	// 1. Типизированная ошибка jwt-go: любой флаг валидации — это 401.
	//    ⚠️ Проверяем ДО текстa: `ValidationError` несёт и `Inner` (ошибка KeyFunc).
	var ve *jwt.ValidationError
	if errors.As(err, &ve) {
		// `ValidationErrorUnverifiable` + Inner = проблема с ключом/подписью —
		// для клиента это всё равно «токен не принят» (401), а не 500.
		return true
	}

	// 2. Текстовая страховка (наши обёртки, старые вызовы).
	return isAuthErrorText(err.Error())
}

// authErrorNeedles — устойчивые фрагменты сообщений jwt-go и нашего кода.
// ⚠️ Регистронезависимо: jwt-go отдаёт «Token is expired by …» с заглавной,
// а обёртки могут менять регистр.
var authErrorNeedles = []string{
	"token is expired",
	"token is malformed",
	"token is unverifiable",
	"signature is invalid",
	"signing method",
	"unexpected signing method",
	"invalid number of segments",
	"token contains an invalid",
	"token is not valid yet",
	"claim not found",
	"token not found",
}

func isAuthErrorText(msg string) bool {
	lower := strings.ToLower(msg)
	for _, needle := range authErrorNeedles {
		if strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

// StatusForError — HTTP-код для ошибки: 401 для проблем с токеном, иначе 500.
func StatusForError(err error) int {
	if IsAuthError(err) {
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}
