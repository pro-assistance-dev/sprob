package fileinfos

import (
	"github.com/pro-assistance/pro-assister/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Download(c *gin.Context) {
	id := c.Param("id")
	item, err := S.Get(c.Request.Context(), id)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	fullPath := F.GetFullPath(&item.FileSystemPath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Download-File-Name", item.OriginalName)
	c.File(*fullPath)
}

func (h *Handler) Create(c *gin.Context) {
	var item models.FileInfo
	files, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = F.Upload(c, &item, files)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = S.Upsert(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}
