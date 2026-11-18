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
//
// ⚠️ ЗАПРОС РАЗБИТ НА ТРИ, и это не оптимизация, а лечение бага.
//
// Раньше здесь было:
//
//	Relation("FormSections.Fields.AnswerVariants")
//	Relation("FormSections.Fields.FieldRules")
//
// в ОДНОМ запросе. Два `has-many` отношения в одном join дают ДЕКАРТОВО
// ПРОИЗВЕДЕНИЕ: bun повторяет строку поля столько раз, сколько у него
// вариантов, на каждое правило. Замер на живых данных: в базе 13 полей у
// формы ярмарки, `GetForm` отдавал 52 записи — каждое поле ×4.
//
// В UI это выглядело как повторяющиеся вопросы в публичной форме и
// удвоенные поля в конструкторе. Клиенты обходили дедупликацией — а это
// значит, что баг приходилось обходить В КАЖДОМ ПРОЕКТЕ.
//
// ⚠️ Поэтому здесь: форма с секциями и полями, затем варианты и правила
// ДВУМЯ запросами, и раскладка по полям в памяти. Три запроса вместо
// одного — дешевле, чем дедупликация в каждом потребителе.
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
		Where("forms.id = ?", formID).
		Scan(c)
	if err != nil {
		return nil, err
	}

	// Плоский список полей — чтобы одним запросом достать варианты и
	// правила и разложить их по местам.
	//
	// ⚠️ `Fields` — срез УКАЗАТЕЛЕЙ (`[]*Field`), поэтому `f` уже указатель.
	// `ID` — `uuid.NullUUID`: у поля может не быть id до сохранения.
	fieldIDs := make([]uuid.UUID, 0, 32)
	byID := make(map[uuid.UUID]*forms.Field, 32)
	for si := range item.FormSections {
		for _, f := range item.FormSections[si].Fields {
			if f == nil || !f.ID.Valid {
				continue
			}
			fieldIDs = append(fieldIDs, f.ID.UUID)
			byID[f.ID.UUID] = f
		}
	}
	if len(fieldIDs) == 0 {
		return &item, nil
	}

	// ⚠️ Два ОТДЕЛЬНЫХ запроса, а не Relation: каждый — свой join, дублей нет.
	variants := []forms.AnswerVariant{}
	if err := r.db.NewSelect().
		Model(&variants).
		Where("field_id IN (?)", bun.In(fieldIDs)).
		Order("answer_variants.item_order").
		Scan(c); err != nil {
		return nil, err
	}
	for i := range variants {
		v := &variants[i]
		if !v.FieldID.Valid {
			continue
		}
		if f, ok := byID[v.FieldID.UUID]; ok {
			// ⚠️ `AnswerVariants` — срез УКАЗАТЕЛЕЙ (`[]*AnswerVariant`).
			f.AnswerVariants = append(f.AnswerVariants, v)
		}
	}

	rules := []forms.FieldRule{}
	if err := r.db.NewSelect().
		Model(&rules).
		Where("field_id IN (?)", bun.In(fieldIDs)).
		Order("field_rules.item_order").
		Scan(c); err != nil {
		return nil, err
	}
	for i := range rules {
		ru := &rules[i]
		if !ru.FieldID.Valid {
			continue
		}
		if f, ok := byID[ru.FieldID.UUID]; ok {
			f.FieldRules = append(f.FieldRules, ru)
		}
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
