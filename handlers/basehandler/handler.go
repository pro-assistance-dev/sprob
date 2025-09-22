package basehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	baseModels "github.com/pro-assistance-dev/sprob/models"
)

func (h *Handler[T]) FTSP(c *gin.Context) {
	data, err := h.S.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, baseModels.FTSPAnswer{Data: data, FTSP: *h.helper.SQL.ExtractFTSP(c.Request.Context())})
}

func (h *Handler[T]) Create(c *gin.Context) {
	var item T
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = h.S.Create(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T]) Get(c *gin.Context) {
	item, err := h.S.Get(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T]) Options(c *gin.Context) {
	label := strcase.ToSnake(c.Param("label"))
	value := strcase.ToSnake(c.Param("value"))
	item, err := h.S.Options(c.Request.Context(), label, value)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T]) GetBySlug(c *gin.Context) {
	item, err := h.S.Get(c.Request.Context(), c.Param("slug"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T]) GetAll(c *gin.Context) {
	items, err := h.S.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler[T]) Delete(c *gin.Context) {
	err := h.S.Delete(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler[T]) Update(c *gin.Context) {
	var item T
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = h.S.Update(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler[T]) UpdateMany(c *gin.Context) {
	var items []*T
	_, err := h.helper.HTTP.GetForm(c, &items)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = h.S.UpdateMany(c.Request.Context(), items)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, items)
}
