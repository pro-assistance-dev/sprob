package answervariants

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"
)

func (r *Repository) Create(c context.Context, item *models.AnswerVariant) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items models.FieldFillsWithCount, err error) {
	items.FieldFills = make(models.FieldFills, 0)
	query := r.helper.DB.IDB(c).NewSelect().Model(&items.FieldFills)
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)
	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.AnswerVariant, error) {
	item := models.AnswerVariant{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.AnswerVariant{}).Where("id = ?", *id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.AnswerVariant) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}
