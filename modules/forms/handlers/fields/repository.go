package fields

import (
	"context"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/middleware"
	"github.com/uptrace/bun"
)

func (r *Repository) Create(c context.Context, item *models.Field) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items models.FieldsWithCount, err error) {
	items.Fields = make(models.Fields, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Fields).
		Relation("FieldFillVariants").
		Relation("ValueType")

	query.Join("join Fields_domains qd on qd.Field_id = Fields.id and qd.domain_id in (?)", bun.In(middleware.ClaimDomainIDS.FromContextSlice(c)))

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.Field, error) {
	item := models.Field{}
	err := r.helper.DB.IDB(c).NewSelect().Model(&item).
		Relation("ValueType").
		Relation("AnswerVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("answer_variants.item_order")
		}).
		Where("?TableAlias.id = ?", id).Scan(c)
	return &item, err
}

func (r *Repository) GetAnthropometryFields(c context.Context) (models.Fields, error) {
	items := make(models.Fields, 0)
	err := r.helper.DB.IDB(c).NewSelect().Model(&items).
		// Where("?TableAlias.code in (?)", bun.In([]string{string(models.AnthropomethryKeyWeight), string(models.AnthropomethryKeyHeight)})).
		Scan(c)
	return items, err
}

func (r *Repository) Delete(c context.Context, id string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.Field{}).Where("id = ?", id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.Field) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) upsertMany(c context.Context, items models.Fields) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().On("conflict (id) do update").
		Model(&items).
		Set(`name = EXCLUDED.name`).
		Set(`short_name = EXCLUDED.short_name`).
		Set(`with_other = EXCLUDED.with_other`).
		Set(`is_files_storage = EXCLUDED.is_files_storage`).
		Set(`col_width = EXCLUDED.col_width`).
		Set(`register_property_order = EXCLUDED.register_property_order`).
		Set(`value_type_id = EXCLUDED.value_type_id`).
		Set(`register_group_id = EXCLUDED.register_group_id`).
		Set(`tag = EXCLUDED.tag`).
		Set(`age_compare = EXCLUDED.age_compare`).
		Exec(c)
	return err
}

func (r *Repository) deleteMany(c context.Context, idPool []uuid.UUID) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().
		Model((*models.Field)(nil)).
		Where("id IN (?)", bun.In(idPool)).
		Exec(c)
	return err
}
