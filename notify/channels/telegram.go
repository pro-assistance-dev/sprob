package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

const tgAPI = "https://api.telegram.org/bot%s/%s"

// Telegram — доставка в Telegram через Bot API. Токен — `TELEGRAM_BOT_TOKEN`.
// chat_id берётся из сообщения (`To`), иначе из `TELEGRAM_NOTIFY_CHAT_ID`, иначе
// ищется по `TELEGRAM_NOTIFY_USERNAME` (боту надо раз написать /start). Найденный
// chat_id кэшируется до перезапуска.
type Telegram struct {
	client *http.Client
}

func NewTelegram() *Telegram {
	return &Telegram{client: &http.Client{Timeout: 15 * time.Second}}
}

func (t *Telegram) Name() string {
	return models.ChannelTelegram
}

func (t *Telegram) Send(ctx context.Context, msg *models.Outbox) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return ErrNotConfigured
	}
	chatID := t.resolveChatID(ctx, token, firstNonEmpty(msg.ToAddress, os.Getenv("TELEGRAM_NOTIFY_CHAT_ID")))
	if chatID == "" {
		return fmt.Errorf("notify/telegram: chat_id не задан (ToAddress/ TELEGRAM_NOTIFY_CHAT_ID/ TELEGRAM_NOTIFY_USERNAME)")
	}
	text := msg.Body
	if msg.Subject != "" {
		text = msg.Subject + "\n\n" + msg.Body
	}
	payload, err := json.Marshal(map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(tgAPI, token, "sendMessage"), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if !out.OK {
		return fmt.Errorf("notify/telegram: %s", out.Description)
	}
	return nil
}

// resolveChatID — chat_id из явного значения, иначе поиск по @username.
func (t *Telegram) resolveChatID(ctx context.Context, token, explicit string) string {
	if v := strings.TrimSpace(explicit); v != "" {
		return v
	}
	username := os.Getenv("TELEGRAM_NOTIFY_USERNAME")
	if username == "" {
		return ""
	}
	return t.resolveChatByUsername(ctx, token, username)
}

// resolveChatByUsername — ищет chat_id по @username в getUpdates.
func (t *Telegram) resolveChatByUsername(ctx context.Context, token, username string) string {
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" {
		return ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(tgAPI, token, "getUpdates"), nil)
	if err != nil {
		return ""
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	var out struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID       int64  `json:"id"`
					Username string `json:"username"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ""
	}
	for i := len(out.Result) - 1; i >= 0; i-- {
		chat := out.Result[i].Message.Chat
		if chat.ID != 0 && strings.EqualFold(chat.Username, username) {
			return strconv.FormatInt(chat.ID, 10)
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
