package ftsppresets

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (r *Repository) Create(c context.Context, item *models.FTSPPreset) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) ([]*models.FTSPPreset, error) {
	items := make([]*models.FTSPPreset, 0)
	err := r.helper.DB.IDB(c).NewSelect().Model(&items).
		Scan(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.FTSPPreset, error) {
	item := models.FTSPPreset{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Where("?TableAlias.id = ?", id).
		Scan(c)
	return &item, err
}

func (r *Repository) Delete(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.FTSPPreset{}).Where("id = ?", id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.FTSPPreset) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}
