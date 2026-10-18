package models

import "errors"

// Ошибки домена — вызывающий код решает, что показать (404/410/500).
var (
	ErrNotFound  = errors.New("link not found")
	ErrDisabled  = errors.New("link disabled")
	ErrExpired   = errors.New("link expired")
	ErrNoDomain  = errors.New("no active link domain")
	ErrBadTarget = errors.New("link target is not resolvable")
)
