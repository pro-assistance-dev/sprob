package notify

import (
	"context"
	"log"
	"os"
	"strings"

	sprobemail "github.com/pro-assistance-dev/sprob/helpers/email"
	"github.com/pro-assistance-dev/sprob/config"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/uptrace/bun"
)

// Notifier — рассылка уведомлений после ответа на публикацию: выбирает включённые
// правила, проверяет условия (AND), рендерит шаблон и отправляет email, записывая
// результат в notification_logs. Отправка — в горутине вызывающего (не блокирует
// ответ респонденту).
type Notifier struct {
	db     *bun.DB
	sender func(to []string, subject, body string) error
}

// New создаёт Notifier со штатным SMTP-отправителем (EMAIL_* из .env).
func New(db *bun.DB) *Notifier {
	return &Notifier{db: db, sender: smtpSend}
}

// SendEmail — отправка письма штатным SMTP (EMAIL_* из окружения).
// Экспортирован для переиспользования (рассылка приглашений).
func SendEmail(to []string, subject, body string) error {
	return smtpSend(to, subject, body)
}

// NotifyByResponse — точка входа после сохранения ответа: находит уведомления
// публикации и отправляет те, чьи условия выполнены. Ошибки логируются, не рвут ответ.
func (n *Notifier) NotifyByResponse(ctx context.Context, publicationID, responseID string) {
	items := make(models.Notifications, 0)
	if err := n.db.NewSelect().
		Model(&items).
		Relation("Rules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("notification_rules.item_order")
		}).
		Where("notifications.publication_id = ?", publicationID).
		Where("notifications.enabled = true").
		Order("notifications.sort_order").
		Scan(ctx); err != nil {
		log.Printf("[notify] select notifications: %v", err)
		return
	}
	if len(items) == 0 {
		return
	}

	values, ruleValues, err := n.responseValues(ctx, responseID)
	if err != nil {
		log.Printf("[notify] load response values: %v", err)
		return
	}

	for _, item := range items {
		n.sendOne(ctx, item, values, ruleValues, publicationID, responseID)
	}

	// Вебхуки — в том же проходе (после писем).
	n.dispatchWebhooks(ctx, publicationID, responseID, values, ruleValues)
}

// sendOne — проверяет условия и отправляет одно уведомление всем получателям.
// Условия сверяются по `ruleValues` (id вариантов/число/строка), а шаблон письма
// рендерится по `values` (читаемые названия вариантов).
func (n *Notifier) sendOne(ctx context.Context, item *models.Notification, values, ruleValues map[string]string, publicationID, responseID string) {
	if !RulesPass(item.Rules, ruleValues) {
		return
	}
	subject := render(item.Subject, values)
	body := render(item.Body, values)
	if strings.TrimSpace(subject) == "" {
		subject = "Новый ответ на опрос"
	}
	for _, addr := range item.Emails {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		status, errText := "sent", ""
		if err := n.sender([]string{addr}, subject, body); err != nil {
			status, errText = "failed", err.Error()
			log.Printf("[notify] send to %s failed: %v", addr, err)
		}
		n.log(ctx, item, publicationID, responseID, addr, subject, status, errText)
	}
}

// responseValues — две карты по кодам полей: текстовые ответы (для шаблона письма,
// radio/set — НАЗВАНИЯ вариантов) и «сырые» (для условий: id вариантов, число,
// строка — то, что хранит правило). Разделяем, потому что правило сравнивает
// id варианта, а в письме нужно название.
func (n *Notifier) responseValues(ctx context.Context, responseID string) (map[string]string, map[string]string, error) {
	response := models.Response{}
	if err := n.db.NewSelect().
		Model(&response).
		Relation("FormFill.FieldFills").
		Relation("FormFill.FieldFills.SelectedAnswerVariants").
		Where("responses.id = ?", responseID).
		Scan(ctx); err != nil {
		return nil, nil, err
	}

	fields, err := n.formFields(ctx, response.PublicationID)
	if err != nil {
		return nil, nil, err
	}
	codeByID := make(map[string]string, len(fields))
	for _, field := range fields {
		codeByID[field.ID.UUID.String()] = field.Code
	}

	values := make(map[string]string)
	ruleValues := make(map[string]string)
	for key, value := range response.Params {
		values["param:"+key] = value
		ruleValues["param:"+key] = value
	}
	if response.FormFill == nil {
		return values, ruleValues, nil
	}
	variantName := n.variantNames(fields)
	for _, fill := range response.FormFill.FieldFills {
		code := codeByID[fill.FieldID.UUID.String()]
		if code == "" {
			continue
		}
		values["field:"+code] = fillText(fill, variantName)
		ruleValues["field:"+code] = fillRawValue(fill)
	}
	return values, ruleValues, nil
}

// log — записывает факт отправки в notification_logs.
func (n *Notifier) log(ctx context.Context, item *models.Notification, publicationID, responseID, addr, subject, status, errText string) {
	row := &models.NotificationLog{
		NotificationID: item.ID,
		PublicationID:  parseUUID(publicationID),
		ResponseID:     parseUUID(responseID),
		ToAddress:      addr,
		Subject:        subject,
		Status:         status,
		Error:          errText,
	}
	if _, err := n.db.NewInsert().Model(row).Exec(ctx); err != nil {
		log.Printf("[notify] log insert: %v", err)
	}
}

// defaultSend — SMTP-отправка по EMAIL_* из окружения.
func smtpSend(to []string, subject, body string) error {
	return sprobemail.NewEmail(sprobemailConfig()).SendEmail(to, subject, body)
}

// sprobemailConfig собирает конфиг отправщика из переменных окружения.
func sprobemailConfig() config.Email {
	return config.Email{
		User:       env("EMAIL_USER"),
		Password:   env("EMAIL_PASSWORD"),
		From:       env("EMAIL_FROM"),
		Server:     env("EMAIL_SERVER"),
		Port:       env("EMAIL_PORT"),
		AuthMethod: env("EMAIL_AUTH_METHOD"),
	}
}

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
