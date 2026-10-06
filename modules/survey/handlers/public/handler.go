package public

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pro-assistance-dev/sprob/helper"
	forms "github.com/pro-assistance-dev/sprob/modules/forms/models"
	"github.com/pro-assistance-dev/sprob/modules/survey/handlers/public/ratelimit"
	"github.com/pro-assistance-dev/sprob/modules/survey/notify"

	"github.com/gin-gonic/gin"
)

// Handler — публичные ручки опросника (без авторизации).
type Handler struct {
	helper *helper.Helper
	s      *Service
	// limiter — защита приёма ответов от флуда/накрутки (по IP+slug).
	limiter *ratelimit.Limiter
}

var H *Handler

// Init собирает цепочку handler→service→repository.
func Init(h *helper.Helper) *Handler {
	db := h.DB.DB
	r := &Repository{db: db}
	s := &Service{r: r, notifier: notify.New(db)}
	handler := &Handler{helper: h, s: s, limiter: ratelimit.New(submitLimit(), submitWindow())}
	H = handler
	return handler
}

// submitLimit/submitWindow — параметры лимита ответов (SURVEY_SUBMIT_LIMIT /
// SURVEY_SUBMIT_WINDOW_SEC), разумные дефолты: 30 ответов за 10 минут с одного IP.
func submitLimit() int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SURVEY_SUBMIT_LIMIT"))); err == nil && v > 0 {
		return v
	}
	return 30
}

func submitWindow() time.Duration {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SURVEY_SUBMIT_WINDOW_SEC"))); err == nil && v > 0 {
		return time.Duration(v) * time.Second
	}
	return 10 * time.Minute
}

// GetForm — GET /api/public/survey/:slug — опубликованная форма + префилл по URL.
func (h *Handler) GetForm(c *gin.Context) {
	slug := c.Param("slug")
	query := queryMap(c)
	res, err := h.s.GetBySlug(c.Request.Context(), slug, query)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Submit — POST /api/public/survey/:slug/responses — принять ответ респондента.
// Тело — multipart/form с полем `form` (как у остальных CRUD sprob).
func (h *Handler) Submit(c *gin.Context) {
	slug := c.Param("slug")

	// Rate-limit по IP+slug: ответы шлются пачкой при накрутке — ограничиваем окно.
	if h.limiter != nil && !h.limiter.Allow(c.ClientIP()+"|"+slug) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "слишком много ответов, попробуйте позже"})
		return
	}

	var payload struct {
		FieldFills []*forms.FieldFill `json:"fieldFills"`
	}
	_, err := h.helper.HTTP.GetForm(c, &payload)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}

	response, err := h.s.SubmitResponse(
		c.Request.Context(),
		slug,
		payload.FieldFills,
		queryMap(c),
		hashSecret(c.ClientIP()),
		"",
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotOpen), errors.Is(err, ErrClosed), errors.Is(err, ErrLimitReached):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInviteRequired), errors.Is(err, ErrInviteInvalid), errors.Is(err, ErrInviteUsed):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	// ⚠️ Ошибки валидации параметров — 400, ДО общего HandleError: тот отдаёт
	// 500 на любое не-nil и делает следующую ветку недостижимой (был баг
	// «missing required param department» → 500 на /api/public/survey/:slug).
	case errors.Is(err, ErrParamRequired), errors.Is(err, ErrParamInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		if h.helper.HTTP.HandleError(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// queryMap собирает query-параметры запроса в map (первое значение на ключ).
func queryMap(c *gin.Context) map[string]string {
	res := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			res[key] = values[0]
		}
	}
	return res
}

// hashSecret — необратимый хэш IP для антифрода (без хранения самого IP).
func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
