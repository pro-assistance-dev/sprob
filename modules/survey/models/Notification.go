package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Триггеры уведомления (когда отправлять).
const (
	// TriggerOnResponse — при каждом ответе на публикацию.
	TriggerOnResponse = "onResponse"
)

// Notification — правило уведомления: КОМУ и ПРИ КАКИХ условиях отправлять письмо
// после ответа на публикацию. Условия — набор NotificationRule (AND): если правил
// нет, письмо уходит при каждом ответе. Это первый шаг (email); Telegram/вебхуки —
// отдельная задача (см. TASKS.md, Фаза 4).
type Notification struct {
	bun.BaseModel `bun:"notifications,alias:notifications"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Publication   *Publication  `bun:"rel:belongs-to" json:"publication"`
	PublicationID uuid.NullUUID `bun:"type:uuid" json:"publicationId"`

	Name string `json:"name"`
	// Emails — список адресов получателей (через запятую в JSON-массиве).
	Emails []string `bun:"emails,type:jsonb" json:"emails"`
	// Subject/Body — шаблон письма; поддерживает плейсхолдеры публикации/ответа
	// (`{{publication}}`, `{{date}}`, `{{param:key}}`, `{{field:code}}`).
	Subject string `bun:"type:text" json:"subject"`
	Body    string `bun:"type:text" json:"body"`

	// Trigger: onResponse (пока единственный).
	Trigger string `json:"trigger"`
	Enabled bool   `json:"enabled"`

	SortOrder int `bun:"sort_order" json:"sortOrder"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`

	Rules NotificationRules `bun:"rel:has-many,join:id=notification_id" json:"rules"`
}

type Notifications []*Notification

type NotificationsWithCount struct {
	Notifications Notifications `json:"items"`
	Count         int           `json:"count"`
}

func (item Notification) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

// BeforeAppendModel проставляет значения по умолчанию.
func (item *Notification) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if item.Trigger == "" {
			item.Trigger = TriggerOnResponse
		}
	}
	return nil
}

// NotificationRule — условие отправки уведомления: ответ на вопрос (FieldCode)
// удовлетворяет правилу (operator + value). Операторы — те же, что у FieldRule.
type NotificationRule struct {
	bun.BaseModel `bun:"notification_rules,alias:notification_rules"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Notification   *Notification `bun:"rel:belongs-to" json:"notification"`
	NotificationID uuid.NullUUID `bun:"type:uuid" json:"notificationId"`

	// FieldCode — код вопроса формы (forms.Field.Code), чей ответ проверяем.
	FieldCode string `json:"fieldCode"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`

	Order uint `bun:"item_order" json:"order"`
}

type NotificationRules []*NotificationRule

type NotificationRulesWithCount struct {
	NotificationRules NotificationRules `json:"items"`
	Count             int               `json:"count"`
}

func (item NotificationRule) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (items NotificationRules) SetIDForChildren(notificationID uuid.NullUUID) {
	for i := range items {
		items[i].NotificationID = notificationID
	}
}

// SetIDForChildren проставляет уведомлению и его правилам id.
func (item *Notification) SetIDForChildren() {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.Rules.SetIDForChildren(item.ID)
}

// NotificationLog — журнал отправок: кому, когда, успех/ошибка. Позволяет
// показать в админке «когда и что ушло» и не терять следы при сбоях SMTP.
type NotificationLog struct {
	bun.BaseModel `bun:"notification_logs,alias:notification_logs"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	NotificationID uuid.NullUUID `bun:"type:uuid" json:"notificationId"`
	PublicationID  uuid.NullUUID `bun:"type:uuid" json:"publicationId"`
	ResponseID     uuid.NullUUID `bun:"type:uuid" json:"responseId"`

	ToAddress string `bun:"to_address" json:"toAddress"`
	Subject   string `json:"subject"`
	Status    string `json:"status"` // sent | failed
	Error     string `bun:"type:text" json:"error"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type NotificationLogs []*NotificationLog

type NotificationLogsWithCount struct {
	NotificationLogs NotificationLogs `json:"items"`
	Count            int              `json:"count"`
}

func (item NotificationLog) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
