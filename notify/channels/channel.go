package channels

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

// Channel — транспорт доставки уведомления. Notifier зовёт его ПОСЛЕ матчинга
// правила, поэтому канал получает уже готовые адрес/тело и не знает про правила.
//
// Реализация обязана быть безопасной для вызова из горутины-воркера: ошибки
// возвращаются наружу (для ретрая через outbox), паники не допускаются.
type Channel interface {
	// Name — совпадает с models.Channel* («email», «telegram», …).
	Name() string
	// Send доставляет сообщение. nil — доставлено.
	Send(ctx context.Context, msg *models.Outbox) error
}
