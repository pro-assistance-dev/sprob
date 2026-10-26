package fileinfos

import (
	"context"
	"mime/multipart"

	"github.com/pro-assistance-dev/sprob/helpers/uploader"
	"github.com/pro-assistance-dev/sprob/models"

	"github.com/gin-gonic/gin"
)

func (s *FilesService) GetFullPath(fileSystemPath *string) *string {
	return s.helper.Uploader.GetFullPath(fileSystemPath)
}

func (s *FilesService) Upload(c *gin.Context, item *models.FileInfo, files map[string][]*multipart.FileHeader) (err error) {
	for i, file := range files {
		if i == item.ID.UUID.String() {
			item.FileSystemPath = uploader.BuildPath(&i)
			err = s.helper.Uploader.Upload(c, file, &item.FileSystemPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PreviousPath — путь файла, который СЕЙЧАС записан в БД у этой записи.
//
// ⚠️ Нужен перед перезаписью: клиент присылает уже новый `fileSystemPath`, и
// прежний файл иначе не найти. Ошибку чтения (записи ещё нет — первый Create)
// считаем «прежнего файла нет»: это не повод валить запрос.
func (s *FilesService) PreviousPath(c context.Context, id string) string {
	if id == "" {
		return ""
	}
	var path string
	err := s.helper.DB.IDB(c).NewSelect().
		Table("file_infos").
		Column("file_system_path").
		Where("id = ?", id).
		Scan(c, &path)
	if err != nil {
		return ""
	}
	return path
}

// RemoveIfReplaced — удалить прежний файл, если путь действительно сменился.
// Сбой удаления не должен валить запрос: запись и новый файл уже на месте.
func (s *FilesService) RemoveIfReplaced(prev *string, current string) {
	if prev == nil || *prev == "" || *prev == current {
		return
	}
	_ = s.helper.Uploader.Remove(prev)
}
