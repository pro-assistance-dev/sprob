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

	"github.com/gin-gonic/gin"
)

type Uploader interface {
	GetUploaderPath() *string
	GetFullPath(*string) *string
	Upload(*gin.Context, []*multipart.FileHeader, *string) error
	Remove(*string) error
	ReadFiles(paths ...string) ([][]byte, error)
}

type LocalUploader struct {
	UploadPath *string
}

func NewLocalUploader(path *string) *LocalUploader {
	staticPath := *path
	return &LocalUploader{
		UploadPath: &staticPath,
	}
}

func (u *LocalUploader) Upload(c *gin.Context, file []*multipart.FileHeader, path *string) (err error) {
	if path == nil {
		return errors.New("file does not relate to anything")
	}
	// ⚠️ `filepath.Join(*UploadPath, *path)` — уже абсолютный и КЛЕЕНЫЙ путь.
	// `filepath.Split` на нём даёт корень `/` в `dirsToFile`, а `os.PathSeparator`
	// приклеивал ещё один `/` → `//app/...`: на Linux `//` = корень хоста
	// контейнера, `mkdir` уходил в `/app` и падал `permission denied` (500).
	dirsToFile, fileName := filepath.Split(filepath.Join(*u.GetUploaderPath(), *path))
	if dirsToFile == "" {
		return errors.New("empty upload dir")
	}
	err = os.MkdirAll(dirsToFile, os.ModePerm)
	if err != nil {
		return err
	}
	err = c.SaveUploadedFile(file[0], filepath.Join(dirsToFile, fileName))
	if err != nil {
		return err
	}
	return nil
}

func (u *LocalUploader) GetUploaderPath() *string {
	return u.UploadPath
}

// Remove удаляет ранее загруженный файл по его ОТНОСИТЕЛЬНОМУ пути
// (`fileSystemPath`), не трогая каталог статики.
//
// ⚠️ Нужен при ПЕРЕЗАПИСИ файла (повторная обрезка фото в галерее): сервер
// присваивает новому файлу новый путь, а старый оставался на диске навсегда —
// «мусорные» файлы копились на каждый кроп. Ошибку отсутствия файла глотаем:
// у старых записей путь мог быть уже почищен.
func (u *LocalUploader) Remove(fileSystemPath *string) error {
	if fileSystemPath == nil || *fileSystemPath == "" {
		return nil
	}
	full := filepath.Join(*u.GetUploaderPath(), *fileSystemPath)
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
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
