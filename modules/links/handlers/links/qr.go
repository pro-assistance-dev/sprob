package links

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance-dev/sprob/modules/links/models"
)

// QrCode — GET /api/links/:id/qr?format=svg|png&size=… — QR обёртки, генерирует
// СЕРВЕР: наклейка печатается независимо от клиента и от его домена.
func (h *Handler) QrCode(c *gin.Context) {
	link, err := h.FindByID(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	url, err := h.ShortURL(c.Request.Context(), link.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.writeQr(c, url)
}

// QrByCode — GET /api/links/qr/:code — QR по коду (печать пакетом, без id).
func (h *Handler) QrByCode(c *gin.Context) {
	link, err := h.FindByCode(c.Request.Context(), c.Param("code"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	url, err := h.ShortURL(c.Request.Context(), link.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.writeQr(c, url)
}

// BatchEnsure — POST /api/links/batch {app, items[]} — пакетный выпуск.
func (h *Handler) BatchEnsure(c *gin.Context) {
	var req struct {
		App   string             `json:"app"`
		Items []models.BatchItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.EnsureBatch(c.Request.Context(), req.App, req.Items)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// SwitchDomain — PUT /api/link-domains/default {host} — сменить активный домен.
// Одна операция на «переехали»: все наклейки начинают вести на новый хост.
func (h *Handler) SwitchDomain(c *gin.Context) {
	var req struct {
		Host string `json:"host"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.switchDomain(c.Request.Context(), req.Host); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"default": req.Host})
}
