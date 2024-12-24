package forms

import (
	"context"
	"fmt"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/pro-assistance-dev/sprob/middleware"

	"github.com/uptrace/bun"
)

func (r *Repository) Create(c context.Context, item *models.Form) (err error) {
	_, err = r.helper.DB.IDB(c).NewInsert().Model(item).Exec(c)
	return err
}

func (r *Repository) GetAll(c context.Context) (items models.FormsWithCount, err error) {
	lib := r.helper.Project.Schemas
	fmt.Println(lib)
	items.Forms = make(models.Forms, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Forms)
	// Relation("Fields.FieldFillVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
	// 	return q.Order("FieldFill_variants.item_order")
	// }).
	// Relation("Fields.FieldExamples").
	// Relation("Fields.ValueType")
	// Relation("Fields.FieldVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
	// 	return q.Order("Field_variants.name")
	// }).
	// Relation("Fields.Children.ValueType").
	// Relation("Fields.Children.FieldFillVariants")
	// Relation("Formulas.FormulaResults")

	// query.Join("join researches_domains on researches_domains.research_id = Forms.id and researches_domains.domain_id in (?)", bun.In(middleware.ClaimDomainIDS.FromContextSlice(c)))
	r.helper.SQL.ExtractFTSP(c).HandleQuery(query)

	items.Count, err = query.ScanAndCount(c)
	return items, err
}

func (r *Repository) Get(c context.Context, id string) (*models.Form, error) {
	item := models.Form{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Relation("FormSections", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("form_sections.item_order")
		}).
		Relation("FormSections.Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("fields.item_order")
		}).
		Relation("FormSections.Fields.ValueType").
		Relation("FormSections.Fields.AnswerVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("answer_variants.item_order")
		}).
		// Relation("Fields.FieldExamples").
		// Relation("Fields.FieldVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
		// 	return q.Order("Field_variants.name")
		// }).
		// Relation("Fields.Children", func(q *bun.SelectQuery) *bun.SelectQuery {
		// 	return q.Order("Fields.item_order")
		// }).
		// Relation("Fields.Children.Children", func(q *bun.SelectQuery) *bun.SelectQuery {
		// 	return q.Order("Fields.item_order")
		// }).
		// Relation("Fields.Children.ValueType").
		// Relation("Fields.Children.Children.ValueType").
		// Relation("Fields.Children.FieldFillVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
		// 	return q.Order("FieldFill_variants.item_order")
		// }).
		// Relation("Formulas.FormulaResults").
		Where("?TableAlias.id = ?", id).Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, err
}

func (r *Repository) Delete(c context.Context, id *string) (err error) {
	_, err = r.helper.DB.IDB(c).NewDelete().Model(&models.Form{}).Where("id = ?", *id).Exec(c)
	return err
}

func (r *Repository) Update(c context.Context, item *models.Form) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Where("id = ?", item.ID).Exec(c)
	return err
}

func (r *Repository) GetForExport(c context.Context, idPool []string) (items models.Forms, err error) {
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("Fields.item_order")
		}).
		Relation("Fields.FieldFillVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("FieldFill_variants.item_order")
		}).
		Relation("Fields.FieldExamples").
		Relation("Fields.ValueType").
		Relation("Fields.FieldVariants", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("Field_variants.name")
		}).
		Relation("Fields.Children.ValueType").
		Relation("Fields.Children.FieldFillVariants").
		Relation("Formulas.FormulaResults")

	if len(idPool) > 0 {
		query = query.Where("?TableAlias.id in (?)", bun.In(idPool))
	}

	query.Join("join researches_domains on researches_domains.research_id = Forms.id and researches_domains.domain_id in (?)", bun.In(middleware.ClaimDomainIDS.FromContextSlice(c)))
	// r.helper.SQL.ExtractFTSP(c).HandleQuery(query)
	err = query.Scan(c)
	return items, err
}

func (r *Repository) UpdateMany(c context.Context, item models.Forms) (err error) {
	_, err = r.helper.DB.IDB(c).NewUpdate().Model(item).Exec(c)
	return err
}
