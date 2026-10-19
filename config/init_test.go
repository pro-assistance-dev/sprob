package config

import (
	"os"
	"path/filepath"
	"runtime"
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

// UPLOAD_PATH из `.env` — главный путь для сервисов на UPLOAD_PATH=/app/static.
func TestProjectSubdirUploadPath(t *testing.T) {
	t.Setenv("UPLOAD_DIR", "")
	t.Setenv("UPLOAD_PATH", "/app/static")
	os.Unsetenv("UPLOAD_DIR")

	if got := projectSubdir("/rdkb/", "x"); got != filepath.Join("/app/static", "x") {
		t.Fatalf("projectSubdir = %q", got)
	}
}

func TestProjectSubdirNameFallback(t *testing.T) {
	t.Setenv("NAME", "")
	os.Unsetenv("NAME")
	if got := nameFallback("/home/u/prog/map"); got != "map" {
		t.Fatalf("nameFallback = %q, want map", got)
	}
}
