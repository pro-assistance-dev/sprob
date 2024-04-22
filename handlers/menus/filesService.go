package menus

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/handlers/basehandler"
)

func (s *FilesService) Upload(c *gin.Context, item basehandler.Filer, files map[string][]*multipart.FileHeader) (err error) {
	for i, file := range files {
		err := s.helper.Uploader.Upload(c, file, item.SetFilePath(&i))
		if err != nil {
			return err
		}
	}
	return nil
}
