package fileinfos

import (
	"os"
	"strings"
	"testing"
)

// ⚠️ Регрессия (26.10): `Handler.Update` НЕ вызывал `FilesService.Upload`, из-за
// чего повторная обрезка фото сохраняла новый путь в БД, но не писала файл →
// `GET /api/static/…` отдавал 404, фото «становилось недоступным».
//
// Тест выше (`TestFilesServiceUpload_Rewrite`) проверяет сам `Upload`, но он НЕ
// поймал бы возврат бага: удаление вызова живёт в `handler.go`. Здесь —
// структурная проверка: в теле `Update` обязан быть вызов `h.f.Upload`.
// ⚠️ Это грубая проверка по исходнику, а не по поведению: полноценный
// handler-тест требует БД (Upsert), а цена ошибки высока — баг был незаметен
// полгода. При рефакторинге заменить на тест с httptest+sqlite.
func TestHandlerUpdateUploadsFile(t *testing.T) {
	src, err := os.ReadFile("handler.go")
	if err != nil {
		t.Fatalf("чтение handler.go: %v", err)
	}
	body := func(name string) string {
		start := strings.Index(string(src), "func (h *Handler) "+name+"(c *gin.Context) {")
		if start < 0 {
			t.Fatalf("не найден метод %s", name)
		}
		rest := string(src)[start:]
		end := strings.Index(rest, "\n}\n")
		if end < 0 {
			t.Fatalf("не найден конец метода %s", name)
		}
		return rest[:end]
	}

	for _, m := range []string{"Create", "Update"} {
		if !strings.Contains(body(m), "h.f.Upload(c, &item, files)") {
			t.Errorf("Handler.%s не вызывает h.f.Upload — файл не будет записан на диск "+
				"(повторная обрезка сделает фото недоступным)", m)
		}
		// `files` должен браться из GetForm, а не выбрасываться в `_`.
		if strings.Contains(body(m), "_, err := h.helper.HTTP.GetForm") {
			t.Errorf("Handler.%s выбрасывает файлы из GetForm (_, err :=) — загрузка не дойдёт", m)
		}
	}
}
