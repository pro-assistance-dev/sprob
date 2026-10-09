package notify

import (
	"context"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/notify/channels"
	"github.com/pro-assistance-dev/sprob/notify/models"

	"github.com/uptrace/bun"
)

const (
	outboxBatch       = 20
	outboxInterval    = 30 * time.Second
	outboxMaxAttempts = 6
)

// Resolver резолвит типы получателей `user`/`role` в адреса конкретного канала
// (email-адрес или telegram chat). Потребитель задаёт его, если хочет адресовать
// уведомления по пользователям/ролям, а не только явными адресами.
type Resolver func(ctx context.Context, targetType, value, channel string) []string

// Notifier — точка отправки уведомлений: принимает события, матчит их с
// правилами из БД, ставит сообщения в outbox и разносит воркером по каналам.
//
// Один на сервис. Потребитель создаёт его из helper (`notify.New(h)`), монтирует
// роуты (`notify.InitRoutes(api, h)`) и публикует события (`Notify.Publish(...)`).
// Отправка НЕ блокирует вызывающий запрос (всё через outbox-воркер).
type Notifier struct {
	db       *bun.DB
	channels map[string]channels.Channel
	resolver Resolver
	autoRun  bool
}

// Option — настройка Notifier (тесты/кастом).
type Option func(*Notifier)

// WithResolver — резолвер получателей `user`/`role`.
func WithResolver(r Resolver) Option {
	return func(n *Notifier) { n.resolver = r }
}

// WithAutoProcess(false) — без фонового воркера (тесты, ручной ProcessDue).
func WithAutoProcess(auto bool) Option {
	return func(n *Notifier) { n.autoRun = auto }
}

// New создаёт Notifier со штатными каналами (email по helper.Email, telegram по
// TELEGRAM_BOT_TOKEN) и запускает outbox-воркер.
func New(h *helper.Helper, opts ...Option) *Notifier {
	n := &Notifier{
		db:      h.DB.DB,
		autoRun: true,
		channels: map[string]channels.Channel{
			models.ChannelEmail:    channels.NewEmail(h.Email),
			models.ChannelTelegram: channels.NewTelegram(),
			models.ChannelWebhook:  channels.NewWebhook(),
		},
	}
	for _, o := range opts {
		o(n)
	}
	if n.autoRun {
		go n.worker()
	}
	return n
}

// Publish — принять событие и разослать по подходящим включённым правилам.
// Возвращается сразу: матчинг и постановка в outbox дешёвые, отправка — в воркере.
func (n *Notifier) Publish(ctx context.Context, event Event) {
	if n == nil {
		return
	}
	if event.At.IsZero() {
		event.At = time.Now()
	}
	rules, err := n.activeRules(ctx, event)
	if err != nil {
		log.Printf("[notify] load rules: %v", err)
		return
	}
	for _, rule := range rules {
		if !rulePasses(rule, event) {
			continue
		}
		n.enqueueRule(ctx, rule, event)
	}
}

// activeRules — включённые правила, чей `Event` совпадает с событием.
func (n *Notifier) activeRules(ctx context.Context, event Event) (models.Rules, error) {
	items := make(models.Rules, 0)
	err := n.db.NewSelect().
		Model(&items).
		Relation("Targets", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_targets.item_order")
		}).
		Relation("Filter", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notify_conds.item_order")
		}).
		Where("notify_rules.enabled = true").
		Order("notify_rules.item_order").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	matched := make(models.Rules, 0, len(items))
	for _, rule := range items {
		if event.Match(rule.Event) {
			matched = append(matched, rule)
		}
	}
	return matched, nil
}

// enqueueRule — кладёт сообщения для одного правила в outbox (по адресу на получателя).
func (n *Notifier) enqueueRule(ctx context.Context, rule *models.Rule, event Event) {
	subject, body := render(rule, event)
	addresses := n.resolveTargets(ctx, rule)
	if len(addresses) == 0 {
		return
	}
	for _, addr := range addresses {
		row := &models.Outbox{
			Channel:       rule.Channel,
			ToAddress:     addr,
			Subject:       subject,
			Body:          body,
			Meta:          rule.Meta,
			RuleID:        rule.ID,
			Event:         event.Name(),
			Entity:        event.Entity,
			ItemID:        event.ItemID,
			NextAttemptAt: time.Now(),
		}
		if _, err := n.db.NewInsert().Model(row).Exec(ctx); err != nil {
			log.Printf("[notify] enqueue to %s: %v", addr, err)
		}
	}
}

// resolveTargets — адреса доставки правила: явные email/telegram/webhook берём
// как есть, user/role — через Resolver.
func (n *Notifier) resolveTargets(ctx context.Context, rule *models.Rule) []string {
	out := make([]string, 0, len(rule.Targets))
	for _, target := range rule.Targets {
		value := strings.TrimSpace(target.Value)
		if value == "" {
			continue
		}
		switch target.Type {
		case models.TargetEmail, models.TargetTelegram, models.TargetWebhook:
			out = append(out, value)
		case models.TargetUser, models.TargetRole:
			if n.resolver == nil {
				log.Printf("[notify] правило %q: получатель %s:%s, но Resolver не задан", rule.Name, target.Type, value)
				continue
			}
			out = append(out, n.resolver(ctx, target.Type, value, rule.Channel)...)
		default:
			log.Printf("[notify] правило %q: неизвестный тип получателя %q", rule.Name, target.Type)
		}
	}
	return out
}

// render — рендер subject/body по снапшоту события ({{.Entity}}, {{.Action}},
// {{.status}}, … — плоские ключи payload; {{.data.Field}} — богатый объект).
// Body со значением «@file:<path>» грузится из шаблона-файла (TEMPLATES_PATH),
// чтобы правила могли переиспользовать существующие gohtml-шаблоны.
func render(rule *models.Rule, event Event) (string, string) {
	data := renderData(event)
	var body string
	if strings.HasPrefix(rule.Body, templateFilePrefix) {
		path := strings.TrimSpace(strings.TrimPrefix(rule.Body, templateFilePrefix))
		if rendered, err := renderFile(path, data); err == nil {
			body = rendered
		} else {
			log.Printf("[notify] шаблон-файл %q: %v", path, err)
		}
	} else {
		body = execTemplate(rule.Body, data)
	}
	return execTemplate(rule.Subject, data), body
}

// renderData — контекст шаблона: плоский снапшот + богатый объект под ключом `data`.
func renderData(event Event) map[string]any {
	out := make(map[string]any, len(event.Payload)+5)
	for k, v := range event.flat() {
		out[k] = v
	}
	if event.Data != nil {
		out["data"] = event.Data
	}
	return out
}

func execTemplate(tpl string, data map[string]any) string {
	if strings.TrimSpace(tpl) == "" {
		return ""
	}
	t, err := template.New("n").Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return tpl // битый шаблон — шлём как есть, не теряем уведомление
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return tpl
	}
	return b.String()
}

// rulePasses — все условия правила (AND). Пусто — проходит.
func rulePasses(rule *models.Rule, event Event) bool {
	if len(rule.Filter) == 0 {
		return true
	}
	data := event.flat()
	for _, cond := range rule.Filter {
		if !condPasses(cond, data) {
			return false
		}
	}
	return true
}

func condPasses(cond *models.Cond, data map[string]string) bool {
	actual := data[cond.Field]
	want := cond.Value
	switch cond.Operator {
	case "equals":
		return actual == want
	case "notEquals":
		return actual != want
	case "contains":
		return strings.Contains(actual, want)
	case "notContains":
		return !strings.Contains(actual, want)
	case "answered":
		return actual != ""
	case "notAnswered":
		return actual == ""
	default:
		return actual == want
	}
}
