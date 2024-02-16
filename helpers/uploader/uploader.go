package uploader

import (
	"crypto/rand"
	"errors"
	"io"
	"log"
	"math/big"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

type Uploader interface {
	GetUploaderPath() *string
	GetFullPath(*string) *string
	Upload(*gin.Context, []*multipart.FileHeader, *string) error
	ReadFiles(paths ...string) ([][]byte, error)
}

type LocalUploader struct {
	UploadPath *string
}

func NewLocalUploader(path *string) *LocalUploader {
	staticPath := filepath.Join(*path)
	return &LocalUploader{
		UploadPath: &staticPath,
	}
}

func (u *LocalUploader) Upload(c *gin.Context, file []*multipart.FileHeader, path *string) (err error) {
	if path == nil {
		return errors.New("file does not relate to anything")
	}
	dirsToFile, fileName := filepath.Split(filepath.Join(*u.GetUploaderPath(), *path))
	if runtime.GOOS == "linux" {
		dirsToFile = string(os.PathSeparator) + dirsToFile
	}
	err = os.MkdirAll(dirsToFile, os.ModePerm)
	if err != nil {
		return err
	}
	err = c.SaveUploadedFile(file[0], filepath.Join([]string{dirsToFile, fileName}...))
	if err != nil {
		return err
	}
	return nil
}

func (u *LocalUploader) GetUploaderPath() *string {
	return u.UploadPath
}

func (u *LocalUploader) GetFullPath(path *string) *string {
	basePath := u.GetUploaderPath()
	fullPath := filepath.Join(*basePath, *path)
	return &fullPath
}

func BuildPath(idFile *string) string {
	fullPath := path.Join(randomString(), randomString(), *idFile)
	return fullPath
}

func (u *LocalUploader) ReadFiles(paths ...string) ([][]byte, error) {
	basePath := u.GetUploaderPath()
	files := make([][]byte, 0)
	for _, path := range paths {
		b, err := u.readFile(filepath.Join(*basePath, path))
		if err != nil {
			return nil, err
		}
		files = append(files, b)
	}
	return files, nil
}

func randomString() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return n.String()
}

func (u *LocalUploader) readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return io.ReadAll(file)
}
