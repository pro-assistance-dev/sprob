package channels

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

const webhookTimeout = 10 * time.Second

// Webhook — внешняя интеграция: POST JSON-тела события на URL (`ToAddress`).
// Если в `Meta["secret"]` задан ключ — тело подписывается HMAC-SHA256
// (`X-Signature: sha256=<hex>`), как у вебхуков опросника.
type Webhook struct {
	client *http.Client
}

func NewWebhook() *Webhook {
	return &Webhook{client: &http.Client{Timeout: webhookTimeout}}
}

func (w *Webhook) Name() string {
	return models.ChannelWebhook
}

func (w *Webhook) Send(ctx context.Context, msg *models.Outbox) error {
	if msg.ToAddress == "" {
		return fmt.Errorf("notify/webhook: пустой URL")
	}
	body := map[string]any{
		"channel": msg.Channel,
		"event":   msg.Event,
		"entity":  msg.Entity,
		"itemId":  msg.ItemID,
		"subject": msg.Subject,
		"body":    msg.Body,
		"meta":    msg.Meta,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, msg.ToAddress, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := msg.Meta["secret"]; secret != "" {
		req.Header.Set("X-Signature", signHMAC(payload, secret))
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("notify/webhook: HTTP %d", resp.StatusCode)
	}
	return nil
}

// signHMAC — HMAC-SHA256 подпись тела: `sha256=<hex>`.
func signHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
