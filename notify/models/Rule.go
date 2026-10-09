package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Каналы доставки уведомления.
const (
	ChannelEmail    = "email"
	ChannelTelegram = "telegram"
	ChannelWebhook  = "webhook"
	ChannelInApp    = "inapp"
)

// Действия события (совпадают с глаголами CRUD + кастомные).
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
)

// Типы получателей правила.
const (
	TargetEmail    = "email"
	TargetTelegram = "telegram"
	TargetUser     = "user"
	TargetRole     = "role"
	TargetWebhook  = "webhook"
)

// Rule — правило уведомления: ЧТО (event) → КУДА (channel) → КОМУ (targets).
// `Event` — «<entity>.<action>» с поддержкой `*` («order.created», «order.*»,
// «*.*»). Пустой `Filter` — слать на каждое событие. Условия — NotificationCond
// (AND), сверяются с плоским снапшотом сущности (payload).
type Rule struct {
	bun.BaseModel `bun:"notify_rules,alias:notify_rules"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Name    string `json:"name"`
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Enabled bool   `json:"enabled"`

	// Subject/Body — шаблоны ({{.Entity}}, {{.Action}}, {{payload.status}}, …).
	Subject string `bun:"type:text" json:"subject"`
	Body    string `bun:"type:text" json:"body"`

	// Специфика канала: webhook — URL+Secret; telegram — переопределение бота.
	Meta Meta `bun:"meta,type:jsonb" json:"meta"`

	Order int `bun:"item_order" json:"order"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`

	Targets Targets `bun:"rel:has-many,join:id=rule_id" json:"targets"`
	Filter  Conds   `bun:"rel:has-many,join:id=rule_id" json:"filter"`
}

type Rules []*Rule

type RulesWithCount struct {
	Rules Rules `json:"items"`
	Count int   `json:"count"`
}

func (item Rule) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (item *Rule) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if _, ok := query.(*bun.InsertQuery); ok {
		if item.Event == "" {
			item.Event = "*.*"
		}
		if item.Channel == "" {
			item.Channel = ChannelEmail
		}
	}
	return nil
}

// SetIDForChildren проставляет правилу и его целям/условиям id.
func (item *Rule) SetIDForChildren() {
	if !item.ID.Valid {
		item.ID = uuid.NullUUID{UUID: uuid.New(), Valid: true}
	}
	item.Targets.SetIDForChildren(item.ID)
	item.Filter.SetIDForChildren(item.ID)
}

func (items Rules) SetIDForChildren() {
	for i := range items {
		items[i].SetIDForChildren()
	}
}

// Target — получатель правила: явный адрес/chat, пользователь или роль.
// Для `user`/`role` значение резолвится в email/telegram получателя на этапе
// отправки (через `Resolver`), поэтому правила не хардкодят адреса.
type Target struct {
	bun.BaseModel `bun:"notify_targets,alias:notify_targets"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Rule   *Rule         `bun:"rel:belongs-to" json:"rule"`
	RuleID uuid.NullUUID `bun:"type:uuid" json:"ruleId"`

	Type  string `json:"type"`
	Value string `json:"value"`

	Order uint `bun:"item_order" json:"order"`
}

type Targets []*Target

func (item Target) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (items Targets) SetIDForChildren(ruleID uuid.NullUUID) {
	for i := range items {
		items[i].RuleID = ruleID
	}
}

// Cond — условие правила по снапшоту сущности (AND). Операторы совпадают с
// FieldRule/NotificationRule форм: equals|notEquals|contains|notContains|… .
type Cond struct {
	bun.BaseModel `bun:"notify_conds,alias:notify_conds"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Rule   *Rule         `bun:"rel:belongs-to" json:"rule"`
	RuleID uuid.NullUUID `bun:"type:uuid" json:"ruleId"`

	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`

	Order uint `bun:"item_order" json:"order"`
}

type Conds []*Cond

func (item Cond) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

func (items Conds) SetIDForChildren(ruleID uuid.NullUUID) {
	for i := range items {
		items[i].RuleID = ruleID
	}
}

// Log — журнал отправок: что ушло, кому, успех/ошибка. Не удаляется правилом,
// чтобы следы отправок не терялись при правке правил.
type Log struct {
	bun.BaseModel `bun:"notify_logs,alias:notify_logs"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	RuleID uuid.NullUUID `bun:"type:uuid" json:"ruleId"`

	Event   string `json:"event"`
	Entity  string `json:"entity"`
	ItemID  string `json:"itemId"`
	Channel string `json:"channel"`

	ToAddress string `bun:"to_address" json:"toAddress"`
	Subject   string `json:"subject"`
	Status    string `json:"status"` // sent | failed
	Error     string `bun:"type:text" json:"error"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type Logs []*Log

type LogsWithCount struct {
	Logs  Logs `json:"items"`
	Count int  `json:"count"`
}

func (item Log) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}
