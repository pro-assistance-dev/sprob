package channels

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

// EmailSender — минимальный контракт отправителя письма (реализует
// `sprob/helpers/email.Email`). Позволяет каналу не зависеть от конфига:
// Notifier берёт уже собранный helper.Email.
type EmailSender interface {
	SendEmail(to []string, subject, body string) error
}

// Email — доставка через SMTP (`sprob/helpers/email`). Без отправителя канал
// «выключен» (dev): Send возвращает ErrNotConfigured.
type Email struct {
	sender EmailSender
}

func NewEmail(sender EmailSender) *Email {
	return &Email{sender: sender}
}

func (e *Email) Name() string {
	return models.ChannelEmail
}

func (e *Email) Send(_ context.Context, msg *models.Outbox) error {
	if e.sender == nil {
		return ErrNotConfigured
	}
	return e.sender.SendEmail([]string{msg.ToAddress}, msg.Subject, msg.Body)
}
