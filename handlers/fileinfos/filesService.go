package fileinfos

import (
	"mime/multipart"
	"pro-assister/helpers/uploader"
	"pro-assister/models"

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
