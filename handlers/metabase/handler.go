package metabase

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func (h *Handler) XLSX(c *gin.Context) {
	cardID := c.Param("cardId")
	url := fmt.Sprintf("/api/card/%s/query/xlsx", cardID)
	file, err := h.helper.Metabase.Request2(url)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	ext := ".xlsx"
	downloadName := time.Now().UTC().Format("data-20060102150405" + ext)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+`"`+downloadName+`"`)
	c.Data(http.StatusOK, "application/octet-stream", file)
}

func (h *Handler) Cards(c *gin.Context) {
	url := "/api/card"
	data, err := h.helper.Metabase.Request2(url)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, data)
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
