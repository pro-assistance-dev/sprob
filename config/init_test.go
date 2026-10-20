package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Прод-`.env` сервисов rdkb: ROOT=/rdkb/, NAME=food. Прежний дефолт давал
// `//rdkb/food/server/static` — `mkdir` в корень контейнера → permission denied.
func TestProjectSubdirLinuxUsesUploadDir(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only default")
	}
	t.Setenv("UPLOAD_DIR", "/app")

	got := projectSubdir("/rdkb/", "static")
	if want := filepath.Join("/app", "static"); got != want {
		t.Fatalf("projectSubdir = %q, want %q", got, want)
	}
	if filepath.Clean(got) != got {
		t.Fatalf("projectSubdir must be a clean path, got %q", got)
	}
}

// UPLOAD_PATH из `.env` — это УЖЕ каталог статики (`/app/static`), `sub` к нему
// не приклеивается: иначе получалось `/app/static/static` (файлы уезжали туда,
// `GET /api/static/<id>` отдавал 404 — прод pros 20.10).
func TestProjectSubdirUploadPath(t *testing.T) {
	t.Setenv("UPLOAD_DIR", "")
	os.Unsetenv("UPLOAD_DIR")
	t.Setenv("UPLOAD_PATH", "/app/static")

	if got := projectSubdir("/rdkb/", "static"); got != "/app/static" {
		t.Fatalf("projectSubdir = %q, want /app/static", got)
	}
	if got := projectSubdir("/rdkb/", "static"); strings.Contains(got, "static/static") {
		t.Fatalf("projectSubdir must not nest static: %q", got)
	}
}

func TestProjectSubdirNameFallback(t *testing.T) {
	t.Setenv("NAME", "")
	os.Unsetenv("NAME")
	if got := nameFallback("/home/u/prog/map"); got != "map" {
		t.Fatalf("nameFallback = %q, want map", got)
	}
}
