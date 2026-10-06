package models

import (
	"context"
	"time"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

// Статусы публикации.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusClosed    = "closed"
)

// Publication — публикация формы во внешний контур: то, чего не хватило в Яндекс
// Формах. Форма (forms.Form) — шаблон вопросов; Publication задаёт КАК и КОМУ она
// отдаётся: ссылку (slug), окно приёма, анонимность, лимит ответов, бренд.
type Publication struct {
	bun.BaseModel `bun:"publications,alias:publications"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Form   *forms.Form   `bun:"rel:belongs-to" json:"form"`
	FormID uuid.NullUUID `bun:"type:uuid" json:"formId"`

	Title string `json:"title"`
	// Slug — публичный идентификатор ссылки: /f/<slug>. Уникален.
	Slug string `json:"slug"`

	// Status: draft | published | closed.
	Status string `json:"status"`

	// Anonym — не требовать приглашение/токен и не связывать ответ с респондентом.
	Anonym bool `json:"anonym"`

	OpensAt  *time.Time `bun:"opens_at" json:"opensAt"`
	ClosesAt *time.Time `bun:"closes_at" json:"closesAt"`
	// Limit — максимальное число ответов (0 — без лимита).
	Limit int `json:"limit"`

	ThankYouText string `bun:"type:text" json:"thankYouText"`
	// Brandbook — бренд публичного рантайма (rdkb/portal/pros/ferma): встраивание
	// в другие проекты без форка кода (бренд-нейтральный рантайм).
	Brandbook string `json:"brandbook"`
	// Theme — тема оформления формы при заполнении (токены навешивает рантайм).
	Theme   *Theme        `bun:"rel:belongs-to" json:"theme"`
	ThemeID uuid.NullUUID `bun:"type:uuid" json:"themeId"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bun:"updated_at" json:"updatedAt"`

	Params    SurveyParams `bun:"rel:has-many,join:id=publication_id" json:"params"`
	Responses Responses    `bun:"rel:has-many,join:id=publication_id" json:"responses"`
}

type Publications []*Publication

type PublicationsWithCount struct {
	Publications Publications `json:"items"`
	Count        int          `json:"count"`
}

func (item Publication) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

// BeforeAppendModel проставляет значения по умолчанию при создании.
// Схема формы создаётся формой-редактором, где пустые поля приходят как "";
// без этого `brandbook`/`status` оставались пустыми строками (не DEFAULT из SQL).
func (item *Publication) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if item.Status == "" {
			item.Status = StatusDraft
		}
		if item.Brandbook == "" {
			item.Brandbook = "rdkb"
		}
	}
	return nil
}
