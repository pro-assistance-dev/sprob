package channels

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pro-assistance-dev/sprob/notify/models"
)

func TestSignHMAC(t *testing.T) {
	payload := []byte(`{"a":1}`)
	got := signHMAC(payload, "s3cr3t")
	mac := hmac.New(sha256.New, []byte("s3cr3t"))
	mac.Write(payload)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("signHMAC = %q, want %q", got, want)
	}
}

func TestWebhookSend(t *testing.T) {
	var gotSig, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Signature")
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		gotBody = string(buf)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := NewWebhook()
	msg := &models.Outbox{
		Channel:   models.ChannelWebhook,
		ToAddress: srv.URL,
		Event:     "order.created",
		Entity:    "orders",
		ItemID:    "42",
		Subject:   "Новый заказ",
		Body:      "тело",
		Meta:      models.Meta{"secret": "k"},
	}
	if err := w.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotSig != signHMAC([]byte(gotBody), "k") {
		t.Errorf("подпись не совпала: %q", gotSig)
	}
	if gotBody == "" {
		t.Error("пустое тело")
	}
}

func TestWebhookHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	w := NewWebhook()
	err := w.Send(context.Background(), &models.Outbox{ToAddress: srv.URL})
	if err == nil {
		t.Fatal("ожидалась ошибка при HTTP 500")
	}
}
