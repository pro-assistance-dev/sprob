package valuetypes

import (
	"context"

	"github.com/pro-assistance-dev/sprob/models"
)

func (r *Repository) Get(c context.Context, id string) (*models.ValueType, error) {
	item := models.ValueType{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		WhereOr("?TableAlias.name = ?", id).
		Scan(c)
	return &item, err
}

func (r *Repository) GetAll(c context.Context) (items models.ValueTypes, err error) {
	items = make(models.ValueTypes, 0)
	err = r.helper.DB.IDB(c).NewSelect().Model(&items).Scan(c)
	return items, err
}
