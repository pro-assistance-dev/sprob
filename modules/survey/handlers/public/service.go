package public

import (
	"context"
	"errors"
	"fmt"
	"time"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/notify"
)

// PublishedForm — публичное представление опубликованной формы: сама форма
// (вопросы), применённые параметры анкеты и значения для префилла полей.
type PublishedForm struct {
	Publication *models.Publication `json:"publication"`
	Form        *forms.Form         `json:"form"`
	// Params — значения параметров из URL (key → value), прошедшие валидацию.
	Params map[string]string `json:"params"`
	// Prefill — значения для предзаполнения полей: код поля → значение.
	Prefill map[string]string `json:"prefill"`
}

// Service — публичный рантайм опросника.
type Service struct {
	r *Repository
	// notifier — рассылка уведомлений после ответа (nil в тестах без БД-рассылки).
	notifier *notify.Notifier
}

// GetBySlug возвращает опубликованную форму по slug, проверяя окно приёма,
// лимит и разрешая параметры анкеты из URL.
// Для неанонимных публикаций обязателен действующий токен приглашения.
func (s *Service) GetBySlug(c context.Context, slug string, query map[string]string) (*PublishedForm, error) {
	publication, err := s.r.GetPublishedBySlug(c, slug)
	if err != nil {
		return nil, err
	}
	if publication == nil {
		return nil, ErrNotFound
	}

	now := time.Now()
	if publication.OpensAt != nil && now.Before(*publication.OpensAt) {
		return nil, ErrNotOpen
	}
	if publication.ClosesAt != nil && now.After(*publication.ClosesAt) {
		return nil, ErrClosed
	}
	if publication.Limit > 0 {
		count, err := s.r.CountResponses(c, publication.ID)
		if err != nil {
			return nil, err
		}
		if count >= publication.Limit {
			return nil, ErrLimitReached
		}
	}

	if !publication.Anonym {
		if _, err := s.requireInvite(c, publication, query["invite"]); err != nil {
			return nil, err
		}
	}

	form, err := s.r.GetForm(c, publication.FormID)
	if err != nil {
		return nil, err
	}

	resolved, prefill, err := resolveParams(publication.Params, query)
	if err != nil {
		return nil, err
	}

	return &PublishedForm{Publication: publication, Form: form, Params: resolved, Prefill: prefill}, nil
}

// requireInvite проверяет токен приглашения: существует, привязан к публикации,
// не использован и не истёк.
func (s *Service) requireInvite(c context.Context, publication *models.Publication, token string) (*models.Invite, error) {
	if token == "" {
		return nil, ErrInviteRequired
	}
	invite, err := s.r.GetInviteByToken(c, token)
	if err != nil {
		return nil, err
	}
	if invite == nil || invite.PublicationID != publication.ID {
		return nil, ErrInviteInvalid
	}
	if !invite.IsUsable() {
		return nil, ErrInviteUsed
	}
	return invite, nil
}

// SubmitResponse принимает ответ респондента: создаёт FormFill/FieldFills и
// обёртку Response с метаданными (параметры, источник, тайминг).
// Для неанонимных публикаций требудет и расходует токен приглашения.
func (s *Service) SubmitResponse(c context.Context, slug string, answers []*forms.FieldFill, query map[string]string, ipHash, respondentHash string) (*models.Response, error) {
	publication, err := s.r.GetPublishedBySlug(c, slug)
	if err != nil {
		return nil, err
	}
	if publication == nil {
		return nil, ErrNotFound
	}
	if publication.Limit > 0 {
		count, err := s.r.CountResponses(c, publication.ID)
		if err != nil {
			return nil, err
		}
		if count >= publication.Limit {
			return nil, ErrLimitReached
		}
	}

	var invite *models.Invite
	if !publication.Anonym {
		invite, err = s.requireInvite(c, publication, query["invite"])
		if err != nil {
			return nil, err
		}
	}

	resolved, _, err := resolveParams(publication.Params, query)
	if err != nil {
		return nil, err
	}

	formFill := &forms.FormFill{FormID: publication.FormID}
	for _, answer := range answers {
		answer.FormFillID = formFill.ID
	}
	formFill.FieldFills = answers

	if err := s.r.CreateFormFill(c, formFill); err != nil {
		return nil, err
	}

	startedAt := time.Now()
	response := &models.Response{
		PublicationID:  publication.ID,
		FormFillID:     formFill.ID,
		Params:         resolved,
		RespondentHash: respondentHash,
		IPHash:         ipHash,
		StartedAt:      startedAt,
		CreatedAt:      time.Now(),
	}
	if err := s.r.CreateResponse(c, response); err != nil {
		return nil, err
	}
	if invite != nil {
		if err := s.r.MarkInviteUsed(c, invite.ID, time.Now()); err != nil {
			return nil, err
		}
	}
	// Уведомления не блокируют ответ респонденту: рассылаем в фоне.
	if s.notifier != nil {
		go s.notifier.NotifyByResponse(context.Background(), publication.ID.UUID.String(), response.ID.UUID.String())
	}
	return response, nil
}

// resolveParams проверяет значения параметров анкеты из URL по их декларации:
// обязательные должны присутствовать, значения — попадать в Options (если заданы).
// Возвращает значения параметров и карту префилла полей (по FieldCode).
func resolveParams(params models.SurveyParams, query map[string]string) (map[string]string, map[string]string, error) {
	resolved := make(map[string]string)
	prefill := make(map[string]string)

	for _, param := range params {
		value := query[param.Key]
		if value == "" {
			if param.Required {
				// ⚠️ Типизированная ошибка: раньше это был обычный fmt.Errorf, и
				// handler отдавал его как 500 ("missing required param department").
				return nil, nil, fmt.Errorf("%w: %q", ErrParamRequired, param.Key)
			}
			continue
		}
		if len(param.Options) > 0 && !contains(param.Options, value) {
			return nil, nil, fmt.Errorf("%w: %q", ErrParamInvalid, param.Key)
		}
		resolved[param.Key] = value
		if param.FieldCode != "" {
			prefill[param.FieldCode] = value
		}
	}
	return resolved, prefill, nil
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

var (
	// Ошибки валидации параметров анкеты (400, а не 500).
	ErrParamRequired = errors.New("missing required param")
	ErrParamInvalid  = errors.New("invalid value for param")

	ErrNotFound       = errors.New("publication not found or not published")
	ErrNotOpen        = errors.New("publication is not open yet")
	ErrClosed         = errors.New("publication is closed")
	ErrLimitReached   = errors.New("response limit reached")
	ErrInviteRequired = errors.New("invite token required")
	ErrInviteInvalid  = errors.New("invite token invalid")
	ErrInviteUsed     = errors.New("invite already used or expired")
)
