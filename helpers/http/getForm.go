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
		return nil, err
	}
	raw := form.Value["form"]
	if len(raw) == 0 {
		return nil, errors.New("multipart field \"form\" is required for non-JSON requests")
	}
	if err = json.Unmarshal([]byte(raw[0]), &item); err != nil {
		return nil, err
	}
	return form.File, nil
}
