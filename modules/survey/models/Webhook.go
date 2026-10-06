package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Webhook — интеграция после ответа: POST JSON на внешний URL с HMAC-подписью.
// Условия те же, что у уведомлений (NotificationRule): вызов идёт только когда
// ответ соответствует всем правилам (AND). Пусто правил — на каждый ответ.
type Webhook struct {
	bun.BaseModel `bun:"webhooks,alias:webhooks"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Publication   *Publication  `bun:"rel:belongs-to" json:"publication"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`

	Name string `json:"name"`
	URL  string `json:"url"`
	// Secret — HMAC-ключ: подпись `X-Signature: sha256=<hex>` от тела.
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`

	SortOrder int `bun:"sort_order" json:"sortOrder"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`

	Rules WebhookRules `bun:"rel:has-many,join:id=webhook_id" json:"rules"`
}

// WebhookRule — условие вызова вебхука (те же операторы, что у NotificationRule).
type WebhookRule struct {
	bun.BaseModel `bun:"webhook_rules,alias:webhook_rules"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Webhook   *Webhook    `bun:"rel:belongs-to" json:"webhook"`
	WebhookID uuid.NullUUID `bun:"type:uuid" json:"webhookId"`

	FieldCode string `json:"fieldCode"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`

	Order uint `bun:"item_order" json:"order"`
}

type WebhookRules []*WebhookRule

type WebhookRulesWithCount struct {
	WebhookRules WebhookRules `json:"items"`
	Count        int          `json:"count"`
}

func (item WebhookRule) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (items WebhookRules) SetIDForChildren(webhookID uuid.NullUUID) {
	for i := range items {
		items[i].WebhookID = webhookID
	}
}

// SetIDForChildren проставляет вебхуку и его правилам id.
func (item *Webhook) SetIDForChildren() {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.Rules.SetIDForChildren(item.ID)
}

// WebhookLog — журнал вызовов (успех/ошибка, код ответа) — как notification_logs.
type WebhookLog struct {
	bun.BaseModel `bun:"webhook_logs,alias:webhook_logs"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	WebhookID     uuid.NullUUID `bun:"type:uuid" json:"webhookId"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`
	ResponseID    uuid.NullUUID `bun:"type:uuid" json:"responseId"`

	URL        string `json:"url"`
	StatusCode int    `bun:"status_code" json:"statusCode"`
	Status     string `json:"status"` // sent | failed
	Error      string `bun:"type:text" json:"error"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type Webhooks []*Webhook

type WebhooksWithCount struct {
	Webhooks Webhooks `json:"items"`
	Count    int      `json:"count"`
}

func (item Webhook) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

type WebhookLogs []*WebhookLog

type WebhookLogsWithCount struct {
	WebhookLogs WebhookLogs `json:"items"`
	Count       int         `json:"count"`
}

func (item WebhookLog) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
