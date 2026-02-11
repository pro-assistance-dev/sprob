package metabase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type MetabaseParameter struct {
	Type   string        `json:"type"`
	Target []interface{} `json:"target"`
	Value  interface{}   `json:"value"`
}

func (h *Handler) XLSX(c *gin.Context) {
	h.Cards()
	name := c.Param("name")
	card := cards.Find(name)
	if card == nil {
		fmt.Println(cards)
		h.helper.HTTP.HandleError(c, fmt.Errorf("card not found"))
		return
	}
	urlVar := fmt.Sprintf("/api/card/%d/query/xlsx", card.ID)

	fmt.Println(c.Request.URL.RawQuery)

	rawQuery := c.Request.URL.RawQuery
	values, err := url.ParseQuery(rawQuery)

	var parameters []MetabaseParameter
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}

		// Пропускаем специальные параметры
		if key == "cardId" {
			continue
		}

		// Создаем параметр Metabase
		param := MetabaseParameter{
			Type: "category", // или определять из параметра
			Target: []interface{}{
				"variable",
				[]interface{}{"template-tag", key},
			},
			Value: vals[0],
		}

		// Если в параметрах указан тип
		if typeKey := fmt.Sprintf("%s_type", key); values.Get(typeKey) != "" {
			param.Type = values.Get(typeKey)
		}

		parameters = append(parameters, param)
	}
	file, err := h.helper.Metabase.Request2(urlVar, parameters)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	ext := ".xlsx"
	downloadName := time.Now().UTC().Format("data-20060102150405" + ext)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+`"`+downloadName+`"`)
	c.Data(http.StatusOK, "application/octet-stream", file)
}

func (h *Handler) Cards() {
	url := "/api/card"
	data, err := h.helper.Metabase.RequestGet(url)
	if err != nil {
		log.Fatal(err)
	}
	// if h.helper.HTTP.HandleError(c, err) {
	// 	return
	// }

	err = json.Unmarshal(data, &cards)
	if err != nil {
		log.Fatal(err)
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
