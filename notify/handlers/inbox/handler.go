package inbox

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/middleware"

	"github.com/gin-gonic/gin"
)

// userIDFromToken — id пользователя из claim `user_id` (пусто, если нет токена).
func userIDFromToken(c *gin.Context) string {
	v, ok := c.Request.Context().Value(middleware.ClaimUserID.String()).(string)
	if !ok {
		return ""
	}
	return v
}

// GetAll — GET /api/notify-inbox[?group=admin]: входящие текущего пользователя.
func (h *Handler) GetAll(c *gin.Context) {
	userID := userIDFromToken(c)
	group := c.Query("group")
	items, err := h.s.GetAll(c.Request.Context(), userID, group)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, items)
}

// Read — POST /api/notify-inbox/:id/read: пометить одно прочитанным.
func (h *Handler) Read(c *gin.Context) {
	if err := h.s.MarkRead(c.Request.Context(), c.Param("id")); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// ReadAll — POST /api/notify-inbox/read-all: пометить всё прочитанным.
func (h *Handler) ReadAll(c *gin.Context) {
	userID := userIDFromToken(c)
	group := c.Query("group")
	if err := h.s.MarkAllRead(c.Request.Context(), userID, group); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
