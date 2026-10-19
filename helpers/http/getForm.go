package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetForm — единый вход для JSON и multipart/form-data (Т1 TECH_DEBT.md sprob).
//
// - Content-Type: application/json — тело парсится в item напрямую, files пустые;
// - multipart/form-data — как раньше: JSON в поле form, файлы из form.File.
//
// Используется во всех CRUD-хендлерах sprob и клиентских проектов (rdkb, portal, pros).
func (i *HTTP) GetForm(c *gin.Context, item interface{}) (map[string][]*multipart.FileHeader, error) {
	if strings.HasPrefix(c.ContentType(), "application/json") {
		if err := json.NewDecoder(c.Request.Body).Decode(&item); err != nil {
			return nil, fmt.Errorf("decode json body: %w", err)
		}
		return map[string][]*multipart.FileHeader{}, nil
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, badForm(err)
	}
	raw := form.Value["form"]
	if len(raw) == 0 {
		return nil, badForm(errors.New("multipart field \"form\" is required for non-JSON requests"))
	}
	if err = json.Unmarshal([]byte(raw[0]), &item); err != nil {
		return nil, badForm(fmt.Errorf("decode form field: %w", err))
	}
	// ⚠️ Файлы ПРИВЯЗАНЫ к записи по `id` (см. `FilesService.Upload`), а клиент мог
	// не прислать `id` в JSON → файл терялся, запись сохранялась без пути
	// (`GET /static/` → 404). Дополняем id из ключа файла и перечитываем тело.
	if fileID := firstFileKey(form.File); fileID != "" && !hasID([]byte(raw[0])) {
		patched := withID([]byte(raw[0]), fileID)
		if err = json.Unmarshal(patched, &item); err != nil {
			return nil, badForm(fmt.Errorf("decode form field: %w", err))
		}
	}
	return form.File, nil
}

// hasID — есть ли в JSON непустой `id`.
func hasID(body []byte) bool {
	var v struct {
		ID string `json:"id"`
	}
	return json.Unmarshal(body, &v) == nil && v.ID != ""
}

// withID — тело с проставленным `id` (остальные поля как были).
func withID(body []byte, id string) []byte {
	var m map[string]json.RawMessage
	if json.Unmarshal(body, &m) != nil {
		return body
	}
	enc, err := json.Marshal(id)
	if err != nil {
		return body
	}
	m["id"] = enc
	out, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return out
}

// firstFileKey — имя первого файла в multipart (у file-infos оно равно id записи).
func firstFileKey(files map[string][]*multipart.FileHeader) string {
	for key := range files {
		return key
	}
	return ""
}

// badForm — ошибка ЗАПРОСА (400, а не 500): тело разобрать нельзя —
// виноват клиент.
func badForm(err error) error {
	return badRequestError{err}
}

type badRequestError struct{ err error }

func (e badRequestError) Error() string  { return e.err.Error() }
func (e badRequestError) Unwrap() error  { return e.err }
func (badRequestError) BadRequest() bool { return true }
