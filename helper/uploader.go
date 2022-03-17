package helper

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type Uploader interface {
	getUploaderPath() *string
	Upload(*gin.Context, *multipart.FileHeader, string) error
}

type LocalUploader struct {
	UploadPath *string
}

func NewLocalUploader(path *string) *LocalUploader {
	return &LocalUploader{
		UploadPath: path,
	}
}

func (u *LocalUploader) Upload(c *gin.Context, file *multipart.FileHeader, name string) (err error) {
	uploadPath := u.getUploaderPath()
	err = c.SaveUploadedFile(file, *uploadPath+"/"+name)
	if err != nil {
		return err
	}
	return nil
}

func (u *LocalUploader) getUploaderPath() *string {
	return u.UploadPath
}
