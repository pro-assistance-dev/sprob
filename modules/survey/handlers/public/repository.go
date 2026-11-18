package public

import (
	"context"
	"database/sql"
	"errors"
	"time"

	basemodels "github.com/pro-assistance-dev/sprob/models"
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
	// ⚠️ ЗАПРОС БЕЗ ВЛОЖЕННЫХ RELATION — и это вынужденно.
	//
	// Даже `Relation("FormSections")` + `Relation("FormSections.Fields")` —
	// это ДВА has-many. bun собирает их join-ом, и строки размножаются:
	// 13 полей у формы ярмарки превращались в 52 ответа (поле ×4).
	// Тот же эффект даёт любая цепочка из двух has-many, даже без
	// AnswerVariants/FieldRules.
	//
	// Поэтому: форма; секции отдельно; поля отдельно; справочник типов
	// отдельно. Четыре запроса по uuid — это доли миллисекунды, а не
	// «оптимизация на потом».
	// ⚠️ `Scan(&item)` без `Relation` НЕ означает «без вложенных сущностей»:
	// bun сам подгружает `rel:has-many` из МОДЕЛИ (`Form.FormSections`,
	// `FormSection.Fields`, `Field.AnswerVariants`). Именно поэтому дубли
	// возникали даже после того, как я убрал `Relation` из запроса — это
	// была не ошибка запроса, а поведение bun по умолчанию.
	//
	// ⚠️ Лечение: `ExcludeColumn` не помогает (это про колонки, не про
	// relation). Правильный путь — ОДИН явный `Relation` на верхнем уровне:
	// `FormSections` без вложенных полей. Дальше поля, варианты и типы
	// грузятся отдельными запросами ниже. Один has-many дублей не даёт.
	err := r.db.NewSelect().
		Model(&item).
		Relation("FormSections", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("form_sections.item_order")
		}).
		Where("forms.id = ?", formID).
		Scan(c)
	if err != nil {
		return nil, err
	}

	// ── Секции ──────────────────────────────────────────────────────
	// Загружены `Relation` выше. ⚠️ Пустой `Fields` у них — норма:
	// поля приходят отдельным запросом (см. ниже), иначе два has-many
	// в одном join снова дадут произведение.
	sections := item.FormSections

	// ⚠️ ОБНУЛЯЕМ поля у секций перед загрузкой.
	//
	// `Relation("FormSections")` тянет `FormSection.Fields` — это ВЛОЖЕННЫЙ
	// has-many, и bun подгружает его сам, потому что relation объявлен в
	// модели (`bun:"rel:has-many"`), а не в запросе. После этого `Scan`
	// собирает join «секции × поля × варианты», и каждое поле приходит
	// несколько раз.
	//
	// ⚠️ `ExcludeColumn` тут не работает: он про колонки, а не про relation.
	// Поэтому просто обнуляем — ниже поля загружаются ОТДЕЛЬНЫМ запросом,
	// и вложение от bun нам не нужно вовсе.
	for i := range sections {
		sections[i].Fields = nil
	}

	// ── Поля ────────────────────────────────────────────────────────
	sectionIDs := make([]uuid.UUID, 0, len(sections))
	secByID := make(map[uuid.UUID]*forms.FormSection, len(sections))
	for _, sec := range sections {
		if sec.ID.Valid {
			sectionIDs = append(sectionIDs, sec.ID.UUID)
			secByID[sec.ID.UUID] = sec
		}
	}
	allFields := forms.Fields{}
	if len(sectionIDs) > 0 {
		if err := r.db.NewSelect().
			Model(&allFields).
			Where("form_section_id IN (?)", bun.In(sectionIDs)).
			Order("fields.item_order").
			Scan(c); err != nil {
			return nil, err
		}
	}

	// ── Справочник типов ────────────────────────────────────────────
	typeIDs := make([]uuid.UUID, 0, 16)
	for _, f := range allFields {
		if f != nil && f.ValueTypeID.Valid {
			typeIDs = append(typeIDs, f.ValueTypeID.UUID)
		}
	}
	if len(typeIDs) > 0 {
		vts := []*basemodels.ValueType{}
		if err := r.db.NewSelect().
			Model(&vts).
			Where("id IN (?)", bun.In(typeIDs)).
			Scan(c); err != nil {
			return nil, err
		}
		vtByID := make(map[uuid.UUID]*basemodels.ValueType, len(vts))
		for _, vt := range vts {
			if vt.ID.Valid {
				vtByID[vt.ID.UUID] = vt
			}
		}
		for _, f := range allFields {
			if f != nil && f.ValueTypeID.Valid {
				if vt, ok := vtByID[f.ValueTypeID.UUID]; ok {
					f.ValueType = vt
				}
			}
		}
	}

	// ── Раскладка полей по секциям ──────────────────────────────────
	for _, f := range allFields {
		if f == nil || !f.FormSectionID.Valid {
			continue
		}
		if sec, ok := secByID[f.FormSectionID.UUID]; ok {
			sec.Fields = append(sec.Fields, f)
		}
	}
	item.FormSections = sections

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
