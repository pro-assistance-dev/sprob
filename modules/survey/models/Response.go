package models

import (
	"time"

	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

// Response — ответ на публикацию: обёртка над forms.FormFill с метаданными,
// нужными для аналитики (источник, параметры анкеты, время прохождения).
// Сами ответы по полям живут в forms.FieldFill — эту модель НЕ дублируем.
type Response struct {
	bun.BaseModel `bun:"responses,alias:responses"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Publication   *Publication  `bun:"rel:belongs-to" json:"publication"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`

	FormFill   *forms.FormFill `bun:"rel:belongs-to" json:"formFill"`
	FormFillID uuid.NullUUID   `bun:"type:uuid" json:"formFillId"`

	// Params — значения параметров анкеты из URL (key → value), например
	// {"department":"cardio"}. Копируются сюда для выгрузок и фильтров,
	// чтобы отчёт не зависел от живого списка SurveyParams.
	Params map[string]string `bun:"params,type:jsonb" json:"params"`

	// RespondentHash — необратимый хэш (email/телефон/id) для защиты от
	// накрутки без хранения персональных данных. Пусто при Anonym-публикации.
	RespondentHash string `bun:"respondent_hash" json:"respondentHash"`
	// IPHash — хэш IP (rate-limit/антифрод), без хранения самого IP.
	IPHash string `bun:"ip_hash" json:"ipHash"`

	StartedAt  time.Time  `bun:"started_at" json:"startedAt"`
	FinishedAt *time.Time `bun:"finished_at" json:"finishedAt"`
	Duration   int        `bun:"duration" json:"duration"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type Responses []*Response

type ResponsesWithCount struct {
	Responses Responses `json:"items"`
	Count     int       `json:"count"`
}

func (item Response) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
