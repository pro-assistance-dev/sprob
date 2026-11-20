package newsfeed

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// ⚠️ Замер 20.11 на прод-контейнере: RSSHub отдаёт «0 байт», затем сразу
// полный документ. Без повтора синк терял ленту целиком.
func TestFetch_RetriesEmptyResponse(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			return // пустой ответ
		}
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rss><channel></channel></rss>`))
	}))
	defer srv.Close()

	f := Fetch{Attempts: 3, RetryPause: 10 * time.Millisecond}
	body, err := f.GetComplete(srv.URL, "rss")
	if err != nil {
		t.Fatalf("должен был пройти со второго раза: %v", err)
	}
	if !isXMLComplete(body, "rss") {
		t.Fatal("документ неполный")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("запросов %d, ожидали 2", got)
	}
}

// Обрезанный документ (без </rss>) — тоже повод повторить.
func TestFetch_RetriesTruncated(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			_, _ = w.Write([]byte(`<rss><channel><item>`)) // оборвано
			return
		}
		_, _ = w.Write([]byte(`<rss><channel></channel></rss>`))
	}))
	defer srv.Close()

	f := Fetch{Attempts: 3, RetryPause: 10 * time.Millisecond}
	if _, err := f.GetComplete(srv.URL, "rss"); err != nil {
		t.Fatalf("должен был пройти со второго раза: %v", err)
	}
}

// Все попытки пустые — внятная ошибка с числом попыток.
func TestFetch_AllAttemptsFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	f := Fetch{Attempts: 2, RetryPause: 10 * time.Millisecond}
	_, err := f.GetComplete(srv.URL, "rss")
	if err == nil {
		t.Fatal("ожидали ошибку")
	}
}
