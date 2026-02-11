package metabase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type MetabaseParameter struct {
	Type   string `json:"type"`
	Target []any  `json:"target"`
	Value  any    `json:"value"`
}

func (h *Handler) XLSX(c *gin.Context) {
	// Обновляем список карточек
	h.Cards()

	name := c.Param("name")
	card := cards.Find(name)
	if card == nil {
		h.helper.HTTP.HandleError(c, fmt.Errorf("card not found"))
		return
	}

	parameters := []map[string]any{}
	for key, vals := range c.Request.URL.Query() {
		if len(vals) == 0 || key == "cardId" {
			continue
		}

		param := map[string]any{
			"type": "string/=", // явно указываем string
			"target": []any{
				"variable",
				[]any{"template-tag", key},
			},
			"value": vals[0],
		}
		parameters = append(parameters, param)
	}

	// Оборачиваем в правильную структуру
	body := map[string]any{
		"parameters": parameters,
	}
	// Выполняем запрос для получения XLSX
	urlPath := fmt.Sprintf("/api/card/%d/query/xlsx", card.ID)
	resp, err := h.helper.Metabase.Post(urlPath, body, nil)
	if err != nil {
		h.helper.HTTP.HandleError(c, err)
		return
	}

	// Отправляем файл
	ext := ".xlsx"
	downloadName := time.Now().UTC().Format("data-20060102150405" + ext)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+`"`+downloadName+`"`)
	c.Data(http.StatusOK, "application/octet-stream", resp.Body)
}

func (h *Handler) Cards() {
	data, err := h.helper.Metabase.Get("/api/card", nil, nil)
	if err != nil {
		// Используем хелпер для логирования ошибки
		log.Println(err)
		return
	}

	err = json.Unmarshal(data.Body, &cards)
	if err != nil {
		log.Println(err)
		return
	}
}

func (h *Handler) Frame(c *gin.Context) {
	questionID := c.Param("questionId")
	claims := jwt.MapClaims{
		"resource": map[string]any{"question": questionID},
		"params":   map[string]any{},
		"exp":      time.Now().Add(100 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, _ := token.SignedString([]byte(h.helper.Metabase.SecretKey))

	iframeURL := fmt.Sprintf(
		"%s/embed/question/%s#bordered=true&titled=true",
		h.helper.Metabase.SiteURL,
		signedToken,
	)

	c.JSON(http.StatusOK, iframeURL)
}
