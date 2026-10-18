package qr

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// alphabet — без визуально спорных символов (l/1/I/O/0): коды читают глазами
// с наклейки, а не только сканером.
const alphabet = "abcdefghijkmnpqrstuvwxyz23456789"

// codeLength 6 → ~10^9 комбинаций: коллизии у 6000 ссылок ничтожны, а код
// остаётся коротким в QR (меньше модулей → надежнее печать).
const codeLength = 6

// NewCode — короткий код обёртки. Проверка на занятость — у сервиса (unique в БД +
// ретрай), генератор коллизии не проверяет.
func NewCode() (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	var b strings.Builder
	b.Grow(codeLength)
	for range codeLength {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String(), nil
}
