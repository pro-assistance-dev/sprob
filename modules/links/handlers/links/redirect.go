package links

import (
	"context"
	"net/http"
	"strconv"

	"github.com/pro-assistance-dev/sprob/modules/links/models"

	"github.com/gin-gonic/gin"
)

// Redirect — публичная обёртка: GET /r/:code → 302 на цель.
// Пользователь не знает о links: страница ошибки минимальна (без бренда сервиса).
func (h *Handler) Redirect(c *gin.Context) {
	ctx := c.Request.Context()
	code := c.Param("code")

	link, err := h.FindByCode(ctx, code)
	if err != nil {
		c.String(http.StatusNotFound, "Ссылка не найдена")
		return
	}
	if err = CheckState(link, now()); err != nil {
		c.String(http.StatusGone, "Ссылка больше не действует")
		return
	}
	target, err := h.ResolveTarget(ctx, link)
	if err != nil || target == "" {
		c.String(http.StatusGone, "Ссылка недоступна")
		return
	}

	// Учёт перехода не задерживает редирект: клиент уже ушёл, ошибка лога — не повод 500.
	ua, ip, ref := c.Request.UserAgent(), c.ClientIP(), c.Request.Referer()
	go func() {
		_ = h.Hit(context.Background(), link, ip, ua, ref)
	}()

	// no-store: иначе браузер/CDN закеширует старый домен и смена не доедет.
	c.Header("Cache-Control", "no-store")
	kind := link.RedirectKind
	if kind != 301 && kind != 302 && kind != 307 && kind != 308 {
		kind = 302
	}
	c.Redirect(kind, target)
}

// Get — GET /api/links/:id.
func (h *Handler) Get(c *gin.Context) {
	var id = c.Param("id")
	link, err := h.FindByID(c.Request.Context(), id)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, link)
}

// Hints — GET /api/links/:id/hits?days=7 — агрегаты переходов (без персональных данных).
func (h *Handler) Hints(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	stats, err := h.HitsByDay(c.Request.Context(), c.Param("id"), days)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": stats, "days": days})
}

// PurgeHitsHandler — DELETE /api/link-hits?days=180 (служебное, links:admin).
func (h *Handler) PurgeHitsHandler(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "180"))
	n, err := h.PurgeHits(c.Request.Context(), days)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": n})
}

// CreateLink — POST /api/link-registry: создаёт ссылку (код генерирует сервер).
// Идемпотентно по (app, targetType, targetId): повтор вернёт существующую —
// массовая печать не должна плодить дубли наклеек с разными кодами.
func (h *Handler) CreateLink(c *gin.Context) {
	link := &models.Link{}
	if _, err := h.helper.HTTP.GetForm(c, link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if link.App == "" || link.TargetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app и targetType обязательны"})
		return
	}
	item, err := h.Ensure(c.Request.Context(), link)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
