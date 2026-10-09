package channels

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"

	"github.com/uptrace/bun"
)

// InApp — доставка «в приложении»: пишет строку в `notify_inbox` для получателя
// (`ToAddress` — id пользователя или имя группы). В отличие от email/telegram
// ничего не отправляет наружу — только хранит до прочтения.
type InApp struct {
	db *bun.DB
}

func NewInApp(db *bun.DB) *InApp {
	return &InApp{db: db}
}

func (c *InApp) Name() string {
	return models.ChannelInApp
}

func (c *InApp) Send(ctx context.Context, msg *models.Outbox) error {
	if c.db == nil {
		return ErrNotConfigured
	}
	row := &models.Inbox{
		Title:  msg.Subject,
		Body:   msg.Body,
		URL:    msg.Meta["url"],
		RuleID: msg.RuleID,
		Event:  msg.Event,
		Entity: msg.Entity,
		ItemID: msg.ItemID,
	}
	// Получатель: numeric/uuid — пользователь; иначе — группа/роль.
	if isUserID(msg.ToAddress) {
		row.UserID = msg.ToAddress
	} else {
		row.Group = msg.ToAddress
	}
	_, err := c.db.NewInsert().Model(row).Exec(ctx)
	return err
}

// isUserID — грубая проверка «похоже на id пользователя» (uuid или числовой sub).
// Пусто/email/телеграм-чат — не пользователь (тогда это группа).
func isUserID(v string) bool {
	if v == "" {
		return false
	}
	for _, r := range v {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || r == '-' {
			continue
		}
		return false
	}
	return true
}
