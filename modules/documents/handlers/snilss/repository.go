package snilss

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/documents/models"
)

func (r *Repository) Create(c context.Context, item *models.Snils) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items models.SnilssWithCount, err error) {
	items.Snilss = make(models.Snilss, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Snilss)

	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.Snils, error) {
	item := models.Snils{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.Snils{}).Where("id = ?", *id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.Snils) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) UpdateMany(c context.Context, item models.Snilss) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Exec(c)
	return err
}
