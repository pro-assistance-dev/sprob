package forms

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/pro-assistance-dev/sprob/middleware"

	"github.com/uptrace/bun"
)

func (r *Repository) Create(c context.Context, item *models.Form) (err error) {
	// Bun v1.2 не каскадирует вложенные has-many при Insert — сохраняем дерево
	// формы (секции → поля → варианты) вручную в одной транзакции.
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

func (r *Repository) GetAll(c context.Context) (items models.FormsWithCount, err error) {
	lib := r.helper.Project.Schemas
	fmt.Println(lib)
	items.Forms = make(models.Forms, 0)
	query := r.helper.DB.IDB(c).NewSelect().
		Model(&items.Forms).
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
		Relation("FormSections.Fields.FieldRules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("field_rules.item_order")
		})
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
		Relation("FormSections.Fields.FieldRules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("field_rules.item_order")
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
	// Обновляем дерево формы целиком: сама форма + вложенные сущности.
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
	// Старое дерево удаляем и пишем новое — простая и надёжная модель для
	// конструктора (перестановка/удаление полей не отслеживается поэлементно).
	if err = r.deleteChildren(c, tx, item.ID); err != nil {
		return err
	}
	if err = r.insertChildren(c, tx, item); err != nil {
		return err
	}
	return tx.Commit()
}

// insertChildren сохраняет секции → поля → варианты ответов формы.
func (r *Repository) insertChildren(c context.Context, tx bun.Tx, item *models.Form) error {
	if len(item.FormSections) == 0 {
		return nil
	}
	if _, err := tx.NewInsert().Model(&item.FormSections).Exec(c); err != nil {
		return err
	}
	fields := make(models.Fields, 0)
	for _, section := range item.FormSections {
		for i := range section.Fields {
			section.Fields[i].FormSectionID = section.ID
		}
		fields = append(fields, section.Fields...)
	}
	if len(fields) > 0 {
		if _, err := tx.NewInsert().Model(&fields).Exec(c); err != nil {
			return err
		}
	}
	variants := make(models.AnswerVariants, 0)
	for _, field := range fields {
		for i := range field.AnswerVariants {
			field.AnswerVariants[i].FieldID = field.ID
		}
		variants = append(variants, field.AnswerVariants...)
	}
	if len(variants) > 0 {
		if _, err := tx.NewInsert().Model(&variants).Exec(c); err != nil {
			return err
		}
	}
	rules := make(models.FieldRules, 0)
	for _, field := range fields {
		for i := range field.FieldRules {
			field.FieldRules[i].FieldID = field.ID
		}
		rules = append(rules, field.FieldRules...)
	}
	if len(rules) > 0 {
		if _, err := tx.NewInsert().Model(&rules).Exec(c); err != nil {
			return err
		}
	}
	return nil
}

// deleteChildren удаляет дерево формы (правила → варианты → поля → секции).
func (r *Repository) deleteChildren(c context.Context, tx bun.Tx, formID uuid.NullUUID) error {
	if _, err := tx.NewDelete().Model((*models.FieldRule)(nil)).
		Where("field_id in (select id from fields where form_section_id in (select id from form_sections where form_id = ?))", formID).Exec(c); err != nil {
		return err
	}
	if _, err := tx.NewDelete().Model((*models.AnswerVariant)(nil)).
		Where("field_id in (select id from fields where form_section_id in (select id from form_sections where form_id = ?))", formID).Exec(c); err != nil {
		return err
	}
	if _, err := tx.NewDelete().Model((*models.Field)(nil)).
		Where("form_section_id in (select id from form_sections where form_id = ?)", formID).Exec(c); err != nil {
		return err
	}
	_, err := tx.NewDelete().Model((*models.FormSection)(nil)).Where("form_id = ?", formID).Exec(c)
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
