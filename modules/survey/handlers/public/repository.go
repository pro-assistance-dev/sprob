package public

import (
	"context"
	"database/sql"
	"errors"
	"time"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Repository — доступ к опубликованным формам и сохранение ответов.
type Repository struct {
	db *bun.DB
}

// GetPublishedBySlug возвращает публикацию по slug, только со статусом published.
func (r *Repository) GetPublishedBySlug(c context.Context, slug string) (*models.Publication, error) {
	item := models.Publication{}
	err := r.db.NewSelect().
		Model(&item).
		Relation("Params", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("survey_params.sort_order")
		}).
		Relation("Theme").
		Where("publications.slug = ?", slug).
		Where("publications.status = ?", models.StatusPublished).
		Scan(c)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetForm возвращает форму с деревом вопросов для публичного рантайма.
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
		Relation("FormSections.Fields.FieldRules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("field_rules.item_order")
		}).
		Where("forms.id = ?", formID).
		Scan(c)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetInviteByToken возвращает приглашение по токену (с публикацией).
func (r *Repository) GetInviteByToken(c context.Context, token string) (*models.Invite, error) {
	item := models.Invite{}
	err := r.db.NewSelect().
		Model(&item).
		Where("invites.token = ?", token).
		Scan(c)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// MarkInviteUsed помечает приглашение использованным (одноразовое).
func (r *Repository) MarkInviteUsed(c context.Context, id uuid.NullUUID, at time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.Invite)(nil)).
		Set("used_at = ?", at).
		Where("id = ?", id).
		Exec(c)
	return err
}

// CountResponses возвращает число ответов по публикации (для лимита).
func (r *Repository) CountResponses(c context.Context, publicationID uuid.NullUUID) (int, error) {
	return r.db.NewSelect().
		Model((*models.Response)(nil)).
		Where("publication_id = ?", publicationID).
		Count(c)
}

// CreateFormFill сохраняет заполнение формы с ответами по полям.
func (r *Repository) CreateFormFill(c context.Context, formFill *forms.FormFill) error {
	formFill.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	if _, err := r.db.NewInsert().Model(formFill).Exec(c); err != nil {
		return err
	}
	if len(formFill.FieldFills) == 0 {
		return nil
	}
	// ⚠️ Bun не каскадирует вложенные has-many — выбранные варианты (для set)
	// сохраняем сами, иначе ответы на «Несколько вариантов» теряются
	// (аналитика/xlsx показывали пустое значение).
	selected := make(forms.SelectedAnswerVariants, 0)
	for i := range formFill.FieldFills {
		fill := formFill.FieldFills[i]
		fill.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
		fill.FormFillID = formFill.ID
		for j := range fill.SelectedAnswerVariants {
			fill.SelectedAnswerVariants[j].FieldFillId = fill.ID
			selected = append(selected, fill.SelectedAnswerVariants[j])
		}
	}
	if _, err := r.db.NewInsert().Model(&formFill.FieldFills).Exec(c); err != nil {
		return err
	}
	if len(selected) > 0 {
		if _, err := r.db.NewInsert().Model(&selected).Exec(c); err != nil {
			return err
		}
	}
	return nil
}

// CreateResponse сохраняет обёртку ответа с метаданными.
func (r *Repository) CreateResponse(c context.Context, response *models.Response) error {
	if !response.ID.Valid {
		response.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	_, err := r.db.NewInsert().Model(response).Exec(c)
	return err
}
