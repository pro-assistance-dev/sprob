package links

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/modules/links/models"
)

// FindByID — ссылка по id (админка).
func (h *Handler) FindByID(ctx context.Context, id string) (*models.Link, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, models.ErrNotFound
	}
	link := &models.Link{}
	err = h.helper.DB.IDB(ctx).NewSelect().Model(link).Where("id = ?", parsed).Scan(ctx)
	if err != nil {
		return nil, models.ErrNotFound
	}
	return link, nil
}

// HitStat — агрегат переходов по дням.
type HitStat struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

// HitsByDay — переходы по ссылке за N дней (без ip/ua: ФЗ-152, только счётчики).
func (h *Handler) HitsByDay(ctx context.Context, id string, days int) ([]HitStat, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, models.ErrNotFound
	}
	stats := make([]HitStat, 0)
	err = h.helper.DB.IDB(ctx).NewSelect().
		TableExpr("link_hits").
		ColumnExpr("to_char(date_trunc('day', created_at), 'YYYY-MM-DD') as day").
		ColumnExpr("count(*)::int as count").
		Where("link_id = ?", parsed).
		Where("created_at >= now() - make_interval(days => ?)", days).
		Group("day").
		Order("day").
		Scan(ctx, &stats)
	return stats, err
}

// EnsureBatch — пакетный выпуск («этаж/здание/отделение/все диеты» одним
// действием, идемпотентно): повтор возвращает существующие ссылки.
func (h *Handler) EnsureBatch(ctx context.Context, app string, items []models.BatchItem) (*models.BatchResult, error) {
	res := &models.BatchResult{Failed: make([]models.BatchItemResult, 0)}
	for _, item := range items {
		link := &models.Link{
			App:        app,
			TargetType: item.TargetType,
			TargetID:   item.TargetID,
			Title:      item.Title,
			Path:       item.Path,
			Enabled:    true,
		}
		existed, err := h.exists(ctx, link)
		if err != nil {
			res.Failed = append(res.Failed, models.BatchItemResult{TargetID: item.TargetID, Error: err.Error()})
			continue
		}
		if existed {
			res.Existed++
			continue
		}
		if err = h.Create(ctx, link); err != nil {
			res.Failed = append(res.Failed, models.BatchItemResult{TargetID: item.TargetID, Error: err.Error()})
			continue
		}
		res.Created++
	}
	return res, nil
}

func (h *Handler) exists(ctx context.Context, link *models.Link) (bool, error) {
	return h.helper.DB.IDB(ctx).NewSelect().Model((*models.Link)(nil)).
		Where("app = ?", link.App).
		Where("target_type = ?", link.TargetType).
		Where("target_id = ?", link.TargetID).
		Exists(ctx)
}

// switchDomain — активный домен обёрток (смена домена = данные, не код).
func (h *Handler) switchDomain(ctx context.Context, host string) error {
	db := h.helper.DB.IDB(ctx)
	if _, err := db.NewUpdate().Model((*models.LinkDomain)(nil)).Set("is_default = false").Where("is_default").Exec(ctx); err != nil {
		return err
	}
	res, err := db.NewUpdate().Model((*models.LinkDomain)(nil)).
		Set("is_default = true").
		Where("host = ?", host).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("домен %q не найден в link_domains", host)
	}
	return nil
}

var _ = time.Now
