package targets

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	basemodels "github.com/pro-assistance-dev/sprob/models"
	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"
	"github.com/uptrace/bun"
)

// Repository — привязка публикаций (форм) к сущностям экосистемы и создание
// «форм по требованию» (регистрация/опрос для события, работника, …).
// (Структура объявлена в init.go; здесь — только методы.)

// ListByTarget — публикации, привязанные к сущности (target_type + target_id).
func (r *Repository) ListByTarget(c context.Context, targetType, targetID string) (models.Publications, error) {
	items := make(models.Publications, 0)
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&items).
		Relation("Form").
		Where("publications.target_type = ?", targetType).
		Where("publications.target_id = ?", targetID).
		Order("publications.created_at").
		Scan(c)
	return items, err
}

// FindByTarget — первая публикация сущности (для идемпотентности ensure).
func (r *Repository) FindByTarget(c context.Context, targetType, targetID string) (*models.Publication, error) {
	item := models.Publication{}
	err := r.helper.DB.IDB(c).NewSelect().
		Model(&item).
		Relation("Form").
		Where("publications.target_type = ?", targetType).
		Where("publications.target_id = ?", targetID).
		Order("publications.created_at").
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

// EnsureForm — создаёт форму (по спецификации полей) + публикацию, привязанную
// к сущности, в одной транзакции. Если публикация сущности уже есть — возвращает
// её (идемпотентно), не создавая дубль. Код поля → id типа значения берётся из
// общей таблицы `value_types` (строка/число/radio/…).
func (r *Repository) EnsureForm(c context.Context, spec FormSpec) (*models.Publication, error) {
	targetID, err := uuid.Parse(spec.TargetID)
	if err != nil {
		return nil, fmt.Errorf("target id не uuid: %w", err)
	}

	if existing, err := r.FindByTarget(c, spec.TargetType, spec.TargetID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	valueTypeIDs, err := r.valueTypeIDs(c)
	if err != nil {
		return nil, err
	}

	tx, err := r.helper.DB.IDB(c).BeginTx(c, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	form := buildForm(spec, valueTypeIDs)
	form.SetIDForChildren()
	if _, err = tx.NewInsert().Model(form).Exec(c); err != nil {
		return nil, err
	}
	if err = insertFormTree(c, tx, form); err != nil {
		return nil, err
	}

	publication := &models.Publication{
		FormID:     form.ID,
		Title:      spec.Title,
		Slug:       slugOrRandom(spec.Slug),
		Status:     models.StatusPublished,
		Anonym:     true,
		Brandbook:  spec.Brandbook,
		TargetType: spec.TargetType,
		TargetID:   uuid.NullUUID{UUID: targetID, Valid: true},
	}
	if publication.Brandbook == "" {
		publication.Brandbook = "rdkb"
	}
	if publication.Title == "" {
		publication.Title = spec.EntityName
	}
	if _, err = tx.NewInsert().Model(publication).Exec(c); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	// Возвращаем с заполненной формой (SQL построил slug/title до маппинга).
	publication.Form = form
	return publication, nil
}

// valueTypeIDs — карта имя типа ответа → id (string/number/radio/…).
func (r *Repository) valueTypeIDs(c context.Context) (map[string]uuid.UUID, error) {
	items := make(basemodels.ValueTypes, 0)
	if err := r.helper.DB.IDB(c).NewSelect().Model(&items).Scan(c); err != nil {
		return nil, err
	}
	out := make(map[string]uuid.UUID, len(items))
	for _, it := range items {
		out[it.Name] = it.ID.UUID
	}
	return out, nil
}

// buildForm — форма по спецификации: одна секция «Анкета», поля по порядку.
func buildForm(spec FormSpec, valueTypeIDs map[string]uuid.UUID) *forms.Form {
	form := &forms.Form{
		ID:     uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Name:   spec.FormName,
		Status: forms.StatusDraft,
	}
	if form.Name == "" {
		form.Name = spec.Title
	}
	section := &forms.FormSection{
		ID:     uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Name:   spec.SectionName,
		FormID: form.ID,
		Order:  0,
	}
	for i, f := range spec.Fields {
		field := &forms.Field{
			ID:            uuid.NullUUID{UUID: uuid.New(), Valid: true},
			Name:          f.Name,
			ShortName:     f.Name,
			Code:          f.Code,
			Order:         uint(i),
			Required:      f.Required,
			FormSectionID: section.ID,
			ValueTypeID:   uuidOrNil(valueTypeIDs[f.ValueType]),
		}
		for j, v := range f.Variants {
			field.AnswerVariants = append(field.AnswerVariants, &forms.AnswerVariant{
				ID:      uuid.NullUUID{UUID: uuid.New(), Valid: true},
				Name:    v,
				Order:   j,
				FieldID: field.ID,
			})
		}
		section.Fields = append(section.Fields, field)
	}
	form.FormSections = append(form.FormSections, section)
	return form
}

// insertFormTree — сохраняет секции → поля → варианты (Bun v1.2 не каскадирует).
func insertFormTree(c context.Context, tx bun.Tx, form *forms.Form) error {
	for _, section := range form.FormSections {
		section.FormID = form.ID
		if _, err := tx.NewInsert().Model(section).Exec(c); err != nil {
			return err
		}
		for _, field := range section.Fields {
			field.FormSectionID = section.ID
			if _, err := tx.NewInsert().Model(field).Exec(c); err != nil {
				return err
			}
			if len(field.AnswerVariants) > 0 {
				if _, err := tx.NewInsert().Model(&field.AnswerVariants).Exec(c); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func uuidOrNil(id uuid.UUID) uuid.NullUUID {
	if id == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

// slugOrRandom — slug из спецификации или короткий случайный (как у create-form).
func slugOrRandom(slug string) string {
	if slug != "" {
		return slug
	}
	return uuid.New().String()[:8]
}
