package reports

import (
	"net/http"
	"strings"

	"github.com/pro-assistance-dev/sprob/helper"

	"github.com/gin-gonic/gin"
)

// Handler — аналитика/выгрузка ответов по публикации (админский контур).
type Handler struct {
	helper *helper.Helper
	s      *Service
}

var H *Handler

// Init собирает цепочку handler→service→repository.
func Init(h *helper.Helper) *Handler {
	r := &Repository{db: h.DB.DB}
	s := &Service{r: r}
	handler := &Handler{helper: h, s: s}
	H = handler
	return handler
}

// Summary — GET /api/publications/:id/summary — сводка по ответам.
func (h *Handler) Summary(c *gin.Context) {
	res, err := h.s.GetSummary(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// Export — GET /api/publications/:id/export?type=xlsx — выгрузка ответов.
func (h *Handler) Export(c *gin.Context) {
	data, name, err := h.s.ExportXlsx(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ByForm — GET /api/publications/by-form/:formId — публикация формы (может быть null).
// Используется табом «Публикация»: у каждой формы — свой бланк.
func (h *Handler) ByForm(c *gin.Context) {
	res, err := h.s.GetPublicationByForm(c.Request.Context(), c.Param("formId"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}

// Responses — GET /api/publications/:id/responses — ответы публикации (таб «Ответы»).
// Фильтр по параметрам анкеты: `?param_<key>=<value>` (например `param_dept=cardio`).
func (h *Handler) Responses(c *gin.Context) {
	filters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) == 0 || !strings.HasPrefix(key, "param_") {
			continue
		}
		filters[strings.TrimPrefix(key, "param_")] = values[0]
	}
	res, err := h.s.GetResponsesByPublication(c.Request.Context(), c.Param("id"), filters)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, res)
}
