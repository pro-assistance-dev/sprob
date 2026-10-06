package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pro-assistance-dev/sprob/modules/survey/models"

	"github.com/uptrace/bun"
)

// webhookTimeout — предел ожидания ответа внешней системы.
const webhookTimeout = 10 * time.Second

// dispatchWebhooks — вызывает включённые вебхуки публикации, чьи условия (AND)
// выполнены. Отправка синхронная (уже в горутине вызывающего), с логом вызова.
func (n *Notifier) dispatchWebhooks(ctx context.Context, publicationID, responseID string, values, ruleValues map[string]string) {
	items := make(models.Webhooks, 0)
	if err := n.db.NewSelect().
		Model(&items).
		Relation("Rules", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("webhook_rules.item_order")
		}).
		Where("webhooks.publication_id = ?", publicationID).
		Where("webhooks.enabled = true").
		Order("webhooks.sort_order").
		Scan(ctx); err != nil {
		log.Printf("[webhook] select: %v", err)
		return
	}
	for _, wh := range items {
		if !webhookRulesPass(wh.Rules, ruleValues) {
			continue
		}
		n.callWebhook(ctx, wh, publicationID, responseID, values)
	}
}

// webhookRulesPass — все правила вебхука выполнены (операторы — как у уведомлений).
func webhookRulesPass(rules models.WebhookRules, values map[string]string) bool {
	for _, rule := range rules {
		conv := &models.NotificationRule{FieldCode: rule.FieldCode, Operator: rule.Operator, Value: rule.Value}
		if !rulePass(conv, values) {
			return false
		}
	}
	return true
}

// callWebhook — POST JSON-тела ответа с HMAC-подписью; результат — в webhook_logs.
func (n *Notifier) callWebhook(ctx context.Context, wh *models.Webhook, publicationID, responseID string, values map[string]string) {
	body := map[string]any{
		"publicationId": publicationID,
		"responseId":    responseID,
		"fields":        values,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		n.logWebhook(ctx, wh, publicationID, responseID, 0, "failed", err.Error())
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(payload))
	if err != nil {
		n.logWebhook(ctx, wh, publicationID, responseID, 0, "failed", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sign(payload, wh.Secret))

	client := &http.Client{Timeout: webhookTimeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[webhook] %s → %s failed: %v", wh.Name, wh.URL, err)
		n.logWebhook(ctx, wh, publicationID, responseID, 0, "failed", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	status := "sent"
	errText := ""
	if resp.StatusCode >= http.StatusBadRequest {
		status = "failed"
		errText = fmt.Sprintf("HTTP %d", resp.StatusCode)
		log.Printf("[webhook] %s → %s HTTP %d", wh.Name, wh.URL, resp.StatusCode)
	}
	n.logWebhook(ctx, wh, publicationID, responseID, resp.StatusCode, status, errText)
}

// sign — HMAC-SHA256 подпись тела: `sha256=<hex>` (секрет из Webhook.Secret).
func sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (n *Notifier) logWebhook(ctx context.Context, wh *models.Webhook, publicationID, responseID string, code int, status, errText string) {
	row := &models.WebhookLog{
		WebhookID:     wh.ID,
		PublicationID: parseUUID(publicationID),
		ResponseID:    parseUUID(responseID),
		URL:           wh.URL,
		StatusCode:    code,
		Status:        status,
		Error:         errText,
	}
	if _, err := n.db.NewInsert().Model(row).Exec(ctx); err != nil {
		log.Printf("[webhook] log insert: %v", err)
	}
}
