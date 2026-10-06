package notifications

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Repository — чтение/замена правил уведомлений публикации.
type Repository struct {
	db *bun.DB
}

// GetAll — уведомления публикации с условиями (в порядке сортировки).
func (r *Repository) GetAll(c context.Context, publicationID string) (models.Notifications, error) {
	items := make(models.Notifications, 0)
	err := r.db.NewSelect().
		Model(&items).
		Relation("Rules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notification_rules.item_order")
		}).
		Where("notifications.publication_id = ?", publicationID).
		Order("notifications.sort_order").
		Scan(c)
	return items, err
}

// GetLogs — журнал отправок публикации (новые сверху).
func (r *Repository) GetLogs(c context.Context, publicationID string) (models.NotificationLogs, error) {
	items := make(models.NotificationLogs, 0)
	err := r.db.NewSelect().
		Model(&items).
		Where("notification_logs.publication_id = ?", publicationID).
		Order("notification_logs.created_at DESC").
		Limit(200).
		Scan(c)
	return items, err
}

// ReplaceAll — атомарно заменяет уведомления публикации: удаляет старые (вместе
// с условиями) и пишет новые. Для редактора это простая и надёжная модель:
// перестановка/удаление не отслеживается поэлементно (как у дерева формы).
func (r *Repository) ReplaceAll(c context.Context, publicationID string, items models.Notifications) error {
	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.deleteAll(c, tx, publicationID); err != nil {
		return err
	}

	pubID := parseUUID(publicationID)
	for i := range items {
		items[i].PublicationID = pubID
		items[i].SetIDForChildren()
	}

	if len(items) > 0 {
		if _, err = tx.NewInsert().Model(&items).Exec(c); err != nil {
			return err
		}
		rules := make(models.NotificationRules, 0)
		for _, item := range items {
			rules = append(rules, item.Rules...)
		}
		if len(rules) > 0 {
			if _, err = tx.NewInsert().Model(&rules).Exec(c); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *Repository) deleteAll(c context.Context, tx bun.Tx, publicationID string) error {
	if _, err := tx.NewDelete().Model((*models.NotificationRule)(nil)).
		Where("notification_id in (select id from notifications where publication_id = ?)", publicationID).
		Exec(c); err != nil {
		return err
	}
	_, err := tx.NewDelete().Model((*models.Notification)(nil)).
		Where("publication_id = ?", publicationID).Exec(c)
	return err
}

func parseUUID(id string) uuid.NullUUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: parsed, Valid: true}
}
