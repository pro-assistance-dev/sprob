package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type testItem struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		var item testItem
		files, err := (&HTTP{}).GetForm(c, &item)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"item": item, "files": len(files)})
	})
	return router
}

// Т1 TECH_DEBT.md sprob: application/json парсится напрямую (без multipart).
func TestGetFormJSON(t *testing.T) {
	router := newTestRouter()
	body, _ := json.Marshal(testItem{Name: "Иван", Age: 30})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Item  testItem `json:"item"`
		Files int      `json:"files"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Item.Name != "Иван" || resp.Item.Age != 30 {
		t.Fatalf("unexpected item: %+v", resp.Item)
	}
	if resp.Files != 0 {
		t.Fatalf("expected 0 files for JSON, got %d", resp.Files)
	}
}

// Обратная совместимость: multipart с полем form + файлы.
func TestGetFormMultipart(t *testing.T) {
	router := newTestRouter()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	raw, _ := json.Marshal(testItem{Name: "Пётр", Age: 40})
	if err := mw.WriteField("form", string(raw)); err != nil {
		t.Fatal(err)
	}
	fw, err := mw.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Item  testItem `json:"item"`
		Files int      `json:"files"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Item.Name != "Пётр" || resp.Item.Age != 40 {
		t.Fatalf("unexpected item: %+v", resp.Item)
	}
	if resp.Files != 1 {
		t.Fatalf("expected 1 file for multipart, got %d", resp.Files)
	}
}

// Multipart без поля form — понятная ошибка, а не panic (index out of range).
func TestGetFormMultipartMissingFormField(t *testing.T) {
	router := newTestRouter()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Fatalf("expected 500 (error), got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "form") {
		t.Fatalf("expected error mentioning form field, got: %s", w.Body.String())
	}
}
