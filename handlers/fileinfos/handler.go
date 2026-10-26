package fileinfos

import (
	"net/http"

	"github.com/pro-assistance-dev/sprob/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Download(c *gin.Context) {
	id := c.Param("id")
	item, err := h.s.Get(c.Request.Context(), id)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	fullPath := h.f.GetFullPath(&item.FileSystemPath)
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
	err = h.f.Upload(c, &item, files)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	err = h.s.Upsert(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Update(c *gin.Context) {
	var item models.FileInfo
	files, err := h.helper.HTTP.GetForm(c, &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	// ⚠️ Старый путь берём из БД ДО загрузки: клиент присылает в JSON уже НОВЫЙ
	// `fileSystemPath` (сгенерированный `FileInfo.setFile`), поэтому из запроса
	// прежний файл не узнать — а без него каждая перезагрузка (повторная обрезка)
	// оставляла бы мусорный файл навсегда.
	prevPath := h.f.PreviousPath(c.Request.Context(), item.ID.UUID.String())

	err = h.f.Upload(c, &item, files)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}

	// ⚠️ Файла в запросе НЕТ (правка подписи/порядка) — путь не мог смениться,
	// а `Upload` его не тронул. Пустой `FileSystemPath` в теле (PATCH-подобные
	// запросы шлют `""`) не должен затереть путь в БД и уж тем более удалить файл.
	if len(files) == 0 && item.FileSystemPath == "" {
		item.FileSystemPath = prevPath
	}

	err = h.s.Upsert(c.Request.Context(), &item)
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	// Удаляем прежний файл только ПОСЛЕ успешного сохранения записи: если что-то
	// сорвалось выше, старое фото остаётся рабочим.
	h.f.RemoveIfReplaced(&prevPath, item.FileSystemPath)
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Delete(c *gin.Context) {
	err := h.s.Delete(c.Request.Context(), c.Param("id"))
	if h.helper.HTTP.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
