package channels

import "errors"

// ErrNotConfigured — канал не настроен (нет креденшелов в окружении). Воркер
// трактует это как «отправлять некуда»: сообщение помечается sent без ошибки,
// чтобы не копить вечные ретраи в dev/без токена.
var ErrNotConfigured = errors.New("notify: channel not configured")
