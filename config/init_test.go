package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTestConfig — smoke-загрузка конфига из минимального test.env
// (во временной папке): не зависит от локальных файлов/окружения, CI-safe.
func TestLoadTestConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.env"), []byte(
		"project.name=test\nproject.root=/tmp\ndb.name=test\n"), 0o600); err != nil {
		t.Fatalf("write test.env: %v", err)
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	if _, err := LoadTestConfig(); err != nil {
		t.Fatalf("LoadTestConfig: %v", err)
	}
}
