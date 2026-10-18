package links

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	qr "github.com/pro-assistance-dev/sprob/modules/links/qr"
)

// qrSize — размер картинки в пикселях, дефолт 512 (печать 3–4 см).
func qrSize(c *gin.Context) int {
	size, _ := strconv.Atoi(c.DefaultQuery("size", "512"))
	if size < 128 || size > 2048 {
		size = 512
	}
	return size
}

func (h *Handler) writeQr(c *gin.Context, url string) {
	m, err := qr.Encode(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	size := qrSize(c)
	if c.DefaultQuery("format", "svg") == "png" {
		c.Data(http.StatusOK, "image/png", m.PNG(size))
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/svg+xml; charset=utf-8", m.SVG(size))
}
