package reports

import (
	"context"
	"database/sql"
	"errors"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Repository — чтение ответов и формы для аналитики/выгрузки.
type Repository struct {
	db *bun.DB
}

// GetPublication возвращает публикацию со связями (форма, параметры).
func (r *Repository) GetPublication(c context.Context, id string) (*models.Publication, error) {
	item := models.Publication{}
	err := r.db.NewSelect().
		Model(&item).
		Relation("Form").
		Relation("Params", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("survey_params.sort_order")
		}).
		Where("publications.id = ?", id).
		Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetForm возвращает форму с деревом вопросов (для колонок и агрегатов).
func (r *Repository) GetForm(c context.Context, formID uuid.NullUUID) (*forms.Form, error) {
	item := forms.Form{}
	err := r.db.NewSelect().
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
		Where("forms.id = ?", formID).
		Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetResponses возвращает ответы публикации с ответами по полям.
// filters — значения параметров анкеты (key → value): оставляем только ответы,
// у которых `params` содержит каждую пару (jsonb: `params->>key = value`).
func (r *Repository) GetResponses(c context.Context, publicationID uuid.NullUUID, filters map[string]string) (models.Responses, error) {
	items := make(models.Responses, 0)
	q := r.db.NewSelect().
		Model(&items).
		Relation("FormFill.FieldFills", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("field_fills.id")
		}).
		Relation("FormFill.FieldFills.SelectedAnswerVariants").
		Where("responses.publication_id = ?", publicationID)
	for key, value := range filters {
		if key == "" || value == "" {
			continue
		}
		// ⚠️ Ключ — ЧЕРЕЗ `bun.Ident` нельзя (это jsonb-путь, не колонка):
		// подставляем как связанный параметр в `->>`, SQL-инъекция исключена.
		q = q.Where("responses.params->>? = ?", key, value)
	}
	err := q.Order("responses.created_at").Scan(c)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetPublicationByForm возвращает публикацию, привязанную к форме (первую).
// Связь «одна форма — одна публикация» (решение 04.10): если публикаций
// несколько (старые данные), отдаём самую свежую.
func (r *Repository) GetPublicationByForm(c context.Context, formID uuid.NullUUID) (*models.Publication, error) {
	item := models.Publication{}
	err := r.db.NewSelect().
		Model(&item).
		Relation("Params", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("survey_params.sort_order")
		}).
		Where("publications.form_id = ?", formID).
		Order("publications.created_at DESC").
		Limit(1).
		Scan(c)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
