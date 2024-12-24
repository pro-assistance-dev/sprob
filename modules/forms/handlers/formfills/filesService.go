package formfills

import (
	"mime/multipart"

	"github.com/pro-assistance-dev/sprob/modules/forms/models"

	"github.com/gin-gonic/gin"
)

func (s *FilesService) Upload(c *gin.Context, item *models.Form, files map[string][]*multipart.FileHeader) (err error) {
	//for i, file := range files {
	//	err := s.helper.Uploader.Upload(c, file, item.SetFilePath(&i))
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}
