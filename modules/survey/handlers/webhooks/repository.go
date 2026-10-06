package webhooks

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Repository — чтение/замена вебхуков публикации.
type Repository struct {
	db *bun.DB
}

// GetAll — вебхуки публикации с условиями (в порядке сортировки).
func (r *Repository) GetAll(c context.Context, publicationID string) (models.Webhooks, error) {
	items := make(models.Webhooks, 0)
	err := r.db.NewSelect().
		Model(&items).
		Relation("Rules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("webhook_rules.item_order")
		}).
		Where("webhooks.publication_id = ?", publicationID).
		Order("webhooks.sort_order").
		Scan(c)
	return items, err
}

// ReplaceAll — атомарно заменяет вебхуки публикации (удаляет старые + пишет новые).
func (r *Repository) ReplaceAll(c context.Context, publicationID string, items models.Webhooks) error {
	tx, err := r.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.NewDelete().Model((*models.WebhookRule)(nil)).
		Where("webhook_id in (select id from webhooks where publication_id = ?)", publicationID).
		Exec(c); err != nil {
		return err
	}
	if _, err = tx.NewDelete().Model((*models.Webhook)(nil)).
		Where("publication_id = ?", publicationID).Exec(c); err != nil {
		return err
	}

	pubID := uuid.NullUUID{UUID: uuid.MustParse(publicationID), Valid: true}
	for i := range items {
		items[i].PublicationID = pubID
		items[i].SetIDForChildren()
	}
	if len(items) > 0 {
		if _, err = tx.NewInsert().Model(&items).Exec(c); err != nil {
			return err
		}
		rules := make(models.WebhookRules, 0)
		for _, wh := range items {
			rules = append(rules, wh.Rules...)
		}
		if len(rules) > 0 {
			if _, err = tx.NewInsert().Model(&rules).Exec(c); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
