package rules

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/notify/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.s.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, models.RulesWithCount{Rules: items.Rules, Count: items.Count})
}

func (h *Handler) FTSP(c *gin.Context) {
	items, err := h.s.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items.Rules, "count": items.Count})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.s.Get(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Create(c *gin.Context) {
	var item models.Rule
	if _, err := h.helper.HTTP.GetForm(c, &item); h.helper.HTTP.HandleError(c, err) {
		return
	}
	if err := h.s.Create(c.Request.Context(), &item); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	var item models.Rule
	if _, err := h.helper.HTTP.GetForm(c, &item); h.helper.HTTP.HandleError(c, err) {
		return
	}
	if err := h.s.Update(c.Request.Context(), &item); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.s.Delete(c.Request.Context(), c.Param("id")); h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Options — для PSelect (id/name) — минимально, чтобы не ломать общий контракт.
func (h *Handler) Options(c *gin.Context) {
	items, err := h.s.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	out := make([]gin.H, 0, len(items.Rules))
	for _, r := range items.Rules {
		out = append(out, gin.H{"id": r.ID, "name": r.Name})
	}
	c.JSON(http.StatusOK, out)
}
