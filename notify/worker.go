package notify

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/pro-assistance-dev/sprob/notify/channels"
	"github.com/pro-assistance-dev/sprob/notify/models"
)

// worker — фоновой разбор очереди: раз в outboxInterval берёт пачку готовых к
// отправке сообщений и рассылает их. Переживает рестарт (очередь в БД).
func (n *Notifier) worker() {
	ticker := time.NewTicker(outboxInterval)
	defer ticker.Stop()
	for range ticker.C {
		n.ProcessDue(context.Background())
	}
}

// ProcessDue — отправить пачку сообщений, у которых наступил next_attempt_at.
// Экспортирован для ручного запуска (тесты, кнопка «повторить»).
//
// Каждое выбранное сообщение СРАЗУ CLAIM'ится (next_attempt_at отодвигается),
// иначе два тика/инстанса могли бы выбрать одну строку и отправить дважды.
func (n *Notifier) ProcessDue(ctx context.Context) {
	items := make(models.OutboxItems, 0)
	err := n.db.NewSelect().
		Model(&items).
		Where("sent_at IS NULL").
		Where("next_attempt_at <= ?", time.Now()).
		Where("attempts < ?", outboxMaxAttempts).
		Order("next_attempt_at").
		Limit(outboxBatch).
		Scan(ctx)
	if err != nil {
		log.Printf("[notify] outbox select: %v", err)
		return
	}
	for _, item := range items {
		if !n.claim(ctx, item) {
			continue // строку уже забрал другой тик/инстанс
		}
		n.deliver(ctx, item)
	}
}

// claim — атомарно «забирает» сообщение, отодвигая next_attempt_at. Возвращает
// false, если строку успел забрать кто-то другой (0 обновлённых).
func (n *Notifier) claim(ctx context.Context, item *models.Outbox) bool {
	res, err := n.db.NewUpdate().Model((*models.Outbox)(nil)).
		Set("next_attempt_at = ?", time.Now().Add(time.Hour)).
		Where("id = ?", item.ID).
		Where("sent_at IS NULL").
		Where("next_attempt_at <= ?", time.Now()).
		Exec(ctx)
	if err != nil {
		log.Printf("[notify] outbox claim: %v", err)
		return false
	}
	n2, err := res.RowsAffected()
	return err == nil && n2 == 1
}

// deliver — одна попытка доставки; обновляет счётчик/ошибку/время следующей попытки.
func (n *Notifier) deliver(ctx context.Context, item *models.Outbox) {
	channel, ok := n.channels[item.Channel]
	if !ok {
		n.finish(ctx, item, "failed", "канал не зарегистрирован: "+item.Channel)
		return
	}
	err := channel.Send(ctx, item)
	switch {
	case err == nil:
		n.finish(ctx, item, "sent", "")
	case errors.Is(err, channels.ErrNotConfigured):
		// Канал выключен (dev/нет токена) — не копим вечные ретраи.
		n.finish(ctx, item, "sent", "канал не настроен — пропущено")
	default:
		item.Attempts++
		item.LastError = err.Error()
		item.NextAttemptAt = time.Now().Add(backoff(item.Attempts))
		if _, uerr := n.db.NewUpdate().Model(item).
			Column("attempts", "last_error", "next_attempt_at").
			Where("id = ?", item.ID).Exec(ctx); uerr != nil {
			log.Printf("[notify] outbox update: %v", uerr)
		}
		log.Printf("[notify] %s → %s failed (attempt %d): %v", item.Channel, item.ToAddress, item.Attempts, err)
	}
}

// finish — пометить сообщение отправленным и записать в журнал.
func (n *Notifier) finish(ctx context.Context, item *models.Outbox, status, errText string) {
	now := time.Now()
	item.SentAt = &now
	item.LastError = errText
	if _, err := n.db.NewUpdate().Model(item).
		Column("sent_at", "last_error").
		Where("id = ?", item.ID).Exec(ctx); err != nil {
		log.Printf("[notify] outbox finish: %v", err)
	}
	logRow := &models.Log{
		RuleID:    item.RuleID,
		Event:     item.Event,
		Entity:    item.Entity,
		ItemID:    item.ItemID,
		Channel:   item.Channel,
		ToAddress: item.ToAddress,
		Subject:   item.Subject,
		Status:    status,
		Error:     errText,
	}
	if _, err := n.db.NewInsert().Model(logRow).Exec(ctx); err != nil {
		log.Printf("[notify] log insert: %v", err)
	}
}

// backoff — экспоненциальная задержка повтора: 1м, 2м, 4м, 8м, 16м.
func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return time.Duration(1<<uint(attempt-1)) * time.Minute
}
