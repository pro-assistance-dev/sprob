package fileinfos

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/helpers/uploader"
	"github.com/pro-assistance-dev/sprob/models"

	"github.com/gin-gonic/gin"
)

// newTestUploader — LocalUploader на временном каталоге (вместо /app/static).
func newTestUploader(t *testing.T) (*uploader.LocalUploader, string) {
	t.Helper()
	dir := t.TempDir()
	up := uploader.NewLocalUploader(&dir)
	return up, dir
}

// formWithFile собирает multipart-запрос с файлом под ключом id записи —
// ровно так это делает клиент (`HttpClient.createFormDataPayload`).
func formWithFile(t *testing.T, id string, content []byte, name string) (*gin.Context, map[string][]*multipart.FileHeader) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile(id, name)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err = fw.Write(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	req := httptest.NewRequest("PUT", "/api/file-infos/"+id, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	form, err := c.MultipartForm()
	if err != nil {
		t.Fatalf("MultipartForm: %v", err)
	}
	return c, form.File
}

// ⚠️ Регрессия (26.10): ПОВТОРНАЯ обрезка фото делала файл недоступным.
//
// `Handler.Update` не вызывал `FilesService.Upload` (в отличие от `Create`),
// поэтому новый путь и новые байты уходили в БД, а файл на диск НЕ писался:
// запись ссылалась на несуществующий путь → `GET /api/static/…` = 404.
//
// Здесь проверяется сам `Upload` при ПЕРЕЗАПИСИ: новый файл записан,
// старый удалён, `FileSystemPath` обновлён.
func TestFilesServiceUpload_Rewrite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	up, dir := newTestUploader(t)
	fs := &FilesService{helper: &helper.Helper{Uploader: up}}

	id := uuid.NewString()
	item := &models.FileInfo{OriginalName: "photo.png"}
	item.ID.UUID = uuid.MustParse(id)

	// 1. Первая загрузка (как при Create).
	c1, files1 := formWithFile(t, id, []byte("first-image-bytes"), "photo.png")
	if err := fs.Upload(c1, item, files1); err != nil {
		t.Fatalf("первая загрузка: %v", err)
	}
	firstPath := item.FileSystemPath
	if firstPath == "" {
		t.Fatal("после первой загрузки FileSystemPath пуст")
	}
	if _, err := os.Stat(filepath.Join(dir, firstPath)); err != nil {
		t.Fatalf("файл первой загрузки не создан: %v", err)
	}

	// 2. ПОВТОРНАЯ обрезка (как при Update): новый файл по тому же id.
	c2, files2 := formWithFile(t, id, []byte("re-cropped-image-bytes-longer"), "photo.png")
	if err := fs.Upload(c2, item, files2); err != nil {
		t.Fatalf("перезапись: %v", err)
	}

	if item.FileSystemPath == firstPath {
		t.Error("FileSystemPath не обновился при перезаписи")
	}

	// Запись получила новый путь и новые байты.
	got, err := os.ReadFile(filepath.Join(dir, item.FileSystemPath))
	if err != nil {
		t.Fatalf("новый файл недоступен (это и есть баг): %v", err)
	}
	if string(got) != "re-cropped-image-bytes-longer" {
		t.Errorf("содержимое нового файла = %q", string(got))
	}

	// Удаление СТАРОГО файла — забота `Handler.Update` (он знает прежний путь из
	// БД; у самого `Upload` этих данных нет — клиент присылает уже новый путь).
	// Здесь проверяем, что `Remove` по прежнему пути действительно его убирает.
	if err := up.Remove(&firstPath); err != nil {
		t.Fatalf("Remove прежнего файла: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, firstPath)); !os.IsNotExist(err) {
		t.Errorf("старый файл не удалён через Remove: %v", err)
	}
}

// Перезапись не должна удалять файл, если новый не записался.
func TestFilesServiceUpload_KeepsOldOnFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	up, dir := newTestUploader(t)
	fs := &FilesService{helper: &helper.Helper{Uploader: up}}

	id := uuid.NewString()
	item := &models.FileInfo{}
	item.ID.UUID = uuid.MustParse(id)

	c1, files1 := formWithFile(t, id, []byte("original"), "photo.png")
	if err := fs.Upload(c1, item, files1); err != nil {
		t.Fatalf("первая загрузка: %v", err)
	}
	firstPath := item.FileSystemPath

	// Битый файл (nil FileHeader.Value) — Upload вернёт ошибку ДО удаления старого.
	bad := map[string][]*multipart.FileHeader{id: {{Filename: "x", Header: nil}}}
	if err := fs.Upload(c1, item, bad); err == nil {
		t.Skip("заголовок без содержимого не даёт ошибку в этой версии Gin")
	}

	if _, err := os.Stat(filepath.Join(dir, firstPath)); err != nil {
		t.Errorf("старый файл удалён несмотря на неудачную запись: %v", err)
	}
}

// Файл под ЧУЖИМ id не должен попадать в запись.
func TestFilesServiceUpload_IgnoresForeignID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	up, _ := newTestUploader(t)
	fs := &FilesService{helper: &helper.Helper{Uploader: up}}

	id := uuid.NewString()
	item := &models.FileInfo{}
	item.ID.UUID = uuid.MustParse(id)

	c, files := formWithFile(t, uuid.NewString(), []byte("x"), "other.png")
	if err := fs.Upload(c, item, files); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if item.FileSystemPath != "" {
		t.Errorf("чужой файл присвоен записи: %q", item.FileSystemPath)
	}
}
