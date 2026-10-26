package fileinfos_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/pro-assistance-dev/sprob/handlers/fileinfos"
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/pro-assistance-dev/sprob/helpers/uploader"
	"github.com/pro-assistance-dev/sprob/models"
	"github.com/pro-assistance-dev/sprob/testkit"

	"github.com/gin-gonic/gin"
)

// newHelper — sqlite-хелпер + таблица file_infos.
//
// ⚠️ testkit.NewSQLiteHelper(models...) не годится: у `models.FileInfo` тег
// `type:uuid,default:uuid_generate_v4()` — это Postgres-функция, sqlite на ней
// падает (`near "(": syntax error`). Поэтому таблицу создаём своим DDL.
func newHelper(t *testing.T) *helper.Helper {
	t.Helper()
	h := testkit.NewSQLiteHelper(t)
	_, err := h.DB.IDB(context.Background()).ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS file_infos (
			id TEXT PRIMARY KEY,
			original_name TEXT NOT NULL DEFAULT '',
			file_system_path TEXT NOT NULL DEFAULT '',
			item_id TEXT,
			item_one_id TEXT,
			description TEXT NOT NULL DEFAULT '',
			item_order INTEGER NOT NULL DEFAULT 0,
			type TEXT NOT NULL DEFAULT ''
		)`)
	if err != nil {
		t.Fatalf("create file_infos: %v", err)
	}
	return h
}

// Integration-тест РЕАЛЬНОГО `Handler.Update` (sqlite через testkit).
//
// ⚠️ Регрессия (26.10): повторная обрезка фото делала файл недоступным —
// `Update` не вызывал `FilesService.Upload`, поэтому в БД уходил новый путь,
// а файла на диске не было (`GET /api/static/…` → 404).
//
// Тест гоняет полный путь: multipart → GetForm → Upload → Upsert(sqlite) и
// проверяет, что файл по НОВОМУ пути реально существует и читается.
func TestHandlerUpdate_WritesFileOnRecrop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHelper(t)
	staticDir := t.TempDir()
	h.Uploader = uploader.NewLocalUploader(&staticDir)
	handler := fileinfos.Init(h)

	id := uuid.NewString()

	// 1. Create — первая загрузка.
	call(t, handler.Create, http.MethodPost, "/api/file-infos", "", id, "first-bytes")
	first := rowPath(t, h, id)
	if _, err := os.Stat(filepath.Join(staticDir, first)); err != nil {
		t.Fatalf("после Create файла нет: %v", err)
	}

	// 2. Update — ПОВТОРНАЯ обрезка (тот же id, новые байты).
	call(t, handler.Update, http.MethodPut, "/api/file-infos/"+id, id, id, "re-cropped-bytes")

	second := rowPath(t, h, id)
	if second == first {
		t.Fatal("Update не обновил file_system_path")
	}
	// Главная проверка: файл по новому пути СУЩЕСТВУЕТ (и это новые байты).
	got, err := os.ReadFile(filepath.Join(staticDir, second))
	if err != nil {
		t.Fatalf("после повторной обрезки файл недоступен: %v", err)
	}
	if string(got) != "re-cropped-bytes" {
		t.Errorf("содержимое = %q, ожидали новые байты", string(got))
	}
	// Старый файл убран (иначе каждая обрезка копит мусор).
	if _, err := os.Stat(filepath.Join(staticDir, first)); !os.IsNotExist(err) {
		t.Errorf("старый файл остался: %v", err)
	}
}

// Update БЕЗ файла (правка подписи) не должен трогать путь и ломать фото.
func TestHandlerUpdate_WithoutFileKeepsPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHelper(t)
	staticDir := t.TempDir()
	h.Uploader = uploader.NewLocalUploader(&staticDir)
	handler := fileinfos.Init(h)

	id := uuid.NewString()
	call(t, handler.Create, http.MethodPost, "/api/file-infos", "", id, "keep-me")
	first := rowPath(t, h, id)

	body, _ := json.Marshal(map[string]any{
		"id": id, "originalName": "photo.png", "fileSystemPath": first, "description": "новая подпись",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/file-infos/"+id, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
	handler.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Update подписи: статус %d, тело %s", w.Code, w.Body.String())
	}
	if got := rowPath(t, h, id); got != first {
		t.Errorf("путь изменился без файла: %q -> %q", first, got)
	}
	if _, err := os.Stat(filepath.Join(staticDir, first)); err != nil {
		t.Errorf("файл пропал после правки подписи: %v", err)
	}
}

// ⚠️ Регрессия (26.10): правка БЕЗ файла, но с ПУСТЫМ `fileSystemPath` в теле.
// Клиенты шлют `""` когда поля нет; если бы такой запрос затирал путь, файл
// оставался на диске, а запись начинала ссылаться в никуда (404) — ровно тот
// же симптом, что и «повторная обрезка ломает файл».
func TestHandlerUpdate_WithoutFileEmptyPathKeepsFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHelper(t)
	staticDir := t.TempDir()
	h.Uploader = uploader.NewLocalUploader(&staticDir)
	handler := fileinfos.Init(h)

	id := uuid.NewString()
	call(t, handler.Create, http.MethodPost, "/api/file-infos", "", id, "keep-me")
	first := rowPath(t, h, id)

	body, _ := json.Marshal(map[string]any{
		"id": id, "originalName": "photo.png", "fileSystemPath": "", "description": "только подпись",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/file-infos/"+id, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id}}
	handler.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("Update: статус %d, тело %s", w.Code, w.Body.String())
	}
	if got := rowPath(t, h, id); got != first {
		t.Errorf("пустой fileSystemPath затёр путь в БД: %q -> %q", first, got)
	}
	if _, err := os.Stat(filepath.Join(staticDir, first)); err != nil {
		t.Errorf("файл удалён из-за пустого пути в теле: %v", err)
	}
}

// call — выполняет handler с multipart-запросом в формате клиента и проверяет 200.
func call(t *testing.T, fn func(*gin.Context), method, target, urlID, fileID, content string) {
	t.Helper()
	c, w := multipartCtx(t, method, target, fileID, content)
	if urlID != "" {
		c.Params = gin.Params{{Key: "id", Value: urlID}}
	}
	fn(c)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: статус %d, тело %s", method, target, w.Code, w.Body.String())
	}
}

// multipartCtx — запрос ровно в том виде, в каком его шлёт клиент:
// поле `form` (JSON записи) + файл под ключом id записи.
func multipartCtx(t *testing.T, method, target, id, content string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	formJSON, _ := json.Marshal(map[string]any{
		"id": id, "originalName": "photo.png", "type": "diet",
	})
	if err := mw.WriteField("form", string(formJSON)); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	fw, err := mw.CreateFormFile(id, "photo.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err = fw.Write([]byte(content)); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err = mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if _, err := c.MultipartForm(); err != nil {
		t.Fatalf("MultipartForm: %v", err)
	}
	return c, w
}

// rowPath — file_system_path записи прямо из БД.
func rowPath(t *testing.T, h *helper.Helper, id string) string {
	t.Helper()
	item := &models.FileInfo{}
	if err := h.DB.IDB(context.Background()).NewSelect().Model(item).Where("id = ?", id).Scan(context.Background()); err != nil {
		t.Fatalf("чтение записи %s: %v", id, err)
	}
	if item.FileSystemPath == "" {
		t.Fatalf("у записи %s пустой file_system_path", id)
	}
	return item.FileSystemPath
}
