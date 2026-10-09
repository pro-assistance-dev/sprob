package targets

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// List — GET /api/publications/by-target/:type/:id — публикации сущности.
func (h *Handler) List(c *gin.Context) {
	items, err := h.s.ListByTarget(c.Request.Context(), c.Param("type"), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// Ensure — POST /api/publications/for-target — создать/вернуть форму сущности.
// Идемпотентно: повторный вызов для той же сущности вернёт существующую публикацию.
func (h *Handler) Ensure(c *gin.Context) {
	var spec FormSpec
	if err := c.ShouldBindJSON(&spec); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.s.EnsureForm(c.Request.Context(), spec)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
