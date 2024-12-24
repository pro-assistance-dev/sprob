package fields

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Create(c *gin.Context) {
	var item models.Field
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}

	err = S.Create(c, &item)

	if h.helper.HTTP.HandleError(c, err) {
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) GetAll(c *gin.Context) {
	items, err := S.GetAll(c.Request.Context())
	if h.helper.HTTP.HandleError(c, err) {
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")
	item, err := S.Get(c, id)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	err := S.Delete(c, id)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) Update(c *gin.Context) {
	var item models.Field
	_, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = S.Update(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
