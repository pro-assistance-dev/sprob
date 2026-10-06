package models

import (
	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

// SurveyParam — параметр анкеты, передаваемый в публичной ссылке
// (`/f/<slug>?department=cardio&doctor=ivanov`). Это ключ параметризации:
// одна и та же форма используется разными отделениями/подразделениями,
// а параметр объясняет, ОТКУДА пришёл ответ (аналитика, ветвление, префилл).
//
// Key — имя в query-строке URL (camelCase, как в Field.Code).
// FieldID — необязательная привязка к полю формы: значение параметра
// предзаполняет это поле (и может быть скрыто от респондента через Hidden).
type SurveyParam struct {
	bun.BaseModel `bun:"survey_params,alias:survey_params"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Publication   *Publication  `bun:"rel:belongs-to" json:"publication"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`

	// Key — имя параметра в URL (например, department).
	Key   string `json:"key"`
	Label string `json:"label"`

	// FieldCode — код поля формы (forms.Field.Code), которое предзаполняется.
	// Пусто — параметр только пишется в метаданные ответа (без поля).
	FieldCode string `json:"fieldCode"`

	// Required — отказ, если параметр не передан в URL.
	Required bool `json:"required"`
	// Hidden — не показывать предзаполненное поле респонденту.
	Hidden bool `json:"hidden"`

	// Options — допустимые значения (jsonb-массив строк). Пусто — любое значение.
	Options []string `bun:"options,type:jsonb" json:"options"`

	SortOrder int `bun:"sort_order" json:"sortOrder"`
}

type SurveyParams []*SurveyParam

type SurveyParamsWithCount struct {
	SurveyParams SurveyParams `json:"items"`
	Count        int          `json:"count"`
}

func (item SurveyParam) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
