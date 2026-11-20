package newsfeed

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// userAgent — браузерный: RSSHub и t.me отдают 403 «боту» без него.
const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// maxBody — потолок ответа: лента канала с картинками доходит до ~60 КБ.
// 16 МБ с запасом и защитой от «бесконечной» ленты.
const maxBody = 16 << 20

// DefaultTimeout — окно ожидания ЗАГОЛОВКОВ. Тело ждём отдельно, окном тишины.
const DefaultTimeout = 30 * time.Second

// Fetch — HTTP-клиент ленты. Нулевое значение готово к работе.
type Fetch struct {
	// Timeout — сколько ждать ЗАГОЛОВКИ (реальный признак «сервер жив»).
	Timeout time.Duration
	// IdleTimeout — сколько ждать следующую порцию ТЕЛА, прежде чем признать,
	// что сервер «затих», и отдать то, что пришло. По умолчанию — 4 с.
	IdleTimeout time.Duration
	// Client — подменить клиент в тестах (httptest).
	Client *http.Client
}

func (f *Fetch) timeout() time.Duration {
	if f.Timeout > 0 {
		return f.Timeout
	}
	return DefaultTimeout
}

func (f *Fetch) idleTimeout() time.Duration {
	if f.IdleTimeout > 0 {
		return f.IdleTimeout
	}
	d := f.timeout() / 6
	if d > 4*time.Second {
		d = 4 * time.Second
	}
	if d < time.Second {
		d = time.Second
	}
	return d
}

func (f *Fetch) client() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ResponseHeaderTimeout: f.timeout(),
			IdleConnTimeout:       30 * time.Second,
			ForceAttemptHTTP2:     true,
		},
		// ⚠️ Client.Timeout НЕ ставим: он рубит и ЧТЕНИЕ тела, то есть даёт
		// `context deadline exceeded` на уже полученном ответе. Это и была
		// грабля 29.10 в darbore («RSSHub недоступен», хотя лента в руках).
	}
}

// Get — GET с браузерными заголовками. Возвращает тело даже при обрыве между
// порциями: неполноту документа проверяет вызывающий (см. `</rss>` в TG).
//
// ⚠️ Почему не `io.ReadAll`: читать до EOF = ждать закрытия соединения.
// RSSHub при обрыве на медиа-вставке отдаёт ~16 КБ и НЕ закрывает — тогда
// `io.ReadAll` висит до общего таймаута (замер 20.11 с сервера: 0.7 с норма,
// 30 с на обрыве). Поэтому читаем порциями с окном тишины.
func (f *Fetch) Get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "ru,en;q=0.8")

	resp, err := f.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return f.readBody(resp.Body)
}

// readBody — чтение порциями с окном тишины.
//
// ⚠️ Канал БУФЕРИЗОВАН: горутина шлёт данные и следом ошибку/EOF. При
// небуферизованном канале второй `Send` блокируется, пока читатель ждёт
// следующую порцию, — и получаем «0 байт» через полный таймаут (поймано на
// ferma 20.11: сервер отдавал всё тело одним `Read` + `EOF`).
func (f *Fetch) readBody(body io.Reader) ([]byte, error) {
	idle := f.idleTimeout()

	type chunk struct {
		data []byte
		err  error
	}
	ch := make(chan chunk, 64)
	go func() {
		got := 0
		tmp := make([]byte, 32<<10)
		for got < maxBody {
			n, rerr := body.Read(tmp)
			if n > 0 {
				got += n
				ch <- chunk{data: append([]byte(nil), tmp[:n]...)}
			}
			if rerr != nil {
				ch <- chunk{err: rerr}
				return
			}
		}
		ch <- chunk{err: io.EOF}
	}()

	var buf []byte
	timer := time.NewTimer(idle)
	defer timer.Stop()
	for len(buf) < maxBody {
		select {
		case c := <-ch:
			if len(c.data) > 0 {
				buf = append(buf, c.data...)
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(idle)
				continue
			}
			// EOF штатный; ошибка после полученных данных — не повод терять
			// ответ (неполноту документа проверяет вызывающий).
			if c.err == nil || c.err == io.EOF || len(buf) > 0 {
				return buf, nil
			}
			return nil, c.err
		case <-timer.C:
			return buf, nil
		}
	}
	return buf, nil
}

// isXMLComplete — документ дошёл целиком (есть закрывающий тег).
// RSSHub при обрыве отдаёт кусок без него — тогда переходим на запасной
// источник, а не отдаём полупустую ленту.
func isXMLComplete(body []byte, root string) bool {
	return bytes.Contains(body, []byte("</"+root+">"))
}
