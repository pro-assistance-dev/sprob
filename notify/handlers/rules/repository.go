package rules

import (
	"context"

	"github.com/pro-assistance-dev/sprob/notify/models"

	"github.com/uptrace/bun"
)

// Create — создаёт правило с получателями и условиями в одной транзакции
// (Bun v1.2 не каскадирует вложенные has-many).
func (r *Repository) Create(c context.Context, item *models.Rule) error {
	item.SetIDForChildren()
	tx, err := r.helper.DB.IDB(c).BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.NewInsert().Model(item).Exec(c); err != nil {
		return err
	}
	if err = r.insertChildren(c, tx, item); err != nil {
		return err
	}
	return tx.Commit()
}

// Update — заменяет правило целиком: строка + пересозданные дети (простая и
// надёжная модель редактора — как у дерева формы).
func (r *Repository) Update(c context.Context, item *models.Rule) error {
	item.SetIDForChildren()
	tx, err := r.helper.DB.IDB(c).BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c); err != nil {
		return err
	}
	if err = r.deleteChildren(c, tx, item.ID); err != nil {
		return err
	}
	if err = r.insertChildren(c, tx, item); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete — удаляет правило (дети уходят по on delete cascade).
func (r *Repository) Delete(c context.Context, id string) error {
	_, err := r.helper.DB.IDB(c).NewDelete().Model((*models.Rule)(nil)).Where("id = ?", id).Exec(c)
	return err
}

// GetAll — правила с получателями и условиями (для админки).
func (r *Repository) GetAll(c context.Context) (models.RulesWithCount, error) {
	items := make(models.Rules, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items).
		Relation("Targets", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_targets.item_order")
		}).
		Relation("Filter", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_conds.item_order")
		}).
		Order("notify_rules.item_order")
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)
	count, err := query.ScanAndCount(c)
	return models.RulesWithCount{Rules: items, Count: count}, err
}

// Get — правило со всеми детьми.
func (r *Repository) Get(c context.Context, id string) (*models.Rule, error) {
	item := models.Rule{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Relation("Targets", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_targets.item_order")
		}).
		Relation("Filter", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_conds.item_order")
		}).
		Where("notify_rules.id = ?", id).
		Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) insertChildren(c context.Context, tx bun.Tx, item *models.Rule) error {
	if len(item.Targets) > 0 {
		if _, err := tx.NewInsert().Model(&item.Targets).Exec(c); err != nil {
			return err
		}
	}
	if len(item.Filter) > 0 {
		if _, err := tx.NewInsert().Model(&item.Filter).Exec(c); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) deleteChildren(c context.Context, tx bun.Tx, ruleID interface{}) error {
	if _, err := tx.NewDelete().Model((*models.Target)(nil)).Where("rule_id = ?", ruleID).Exec(c); err != nil {
		return err
	}
	_, err := tx.NewDelete().Model((*models.Cond)(nil)).Where("rule_id = ?", ruleID).Exec(c)
	return err
}
