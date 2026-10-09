package notify

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// templateFilePrefix — маркер в Rule.Body «рендерить из файла»: `@file:email/order_new.gohtml`.
// Путь — относительно TEMPLATES_PATH (или /app/templates), чтобы правила
// переиспользовали существующие gohtml-шаблоны проекта.
const templateFilePrefix = "@file:"

// renderFile — рендер шаблона-файла с общими _header/_footer (как в проектах).
func renderFile(relPath string, data map[string]any) (string, error) {
	dir := strings.TrimSpace(os.Getenv("TEMPLATES_PATH"))
	if dir == "" {
		dir = "/app/templates"
	}
	full := filepath.Join(dir, relPath)
	files := []string{full}
	for _, shared := range []string{"_header.html", "_footer.html"} {
		p := filepath.Join(dir, shared)
		if _, err := os.Stat(p); err == nil {
			files = append(files, p)
		}
	}
	t, err := template.New(filepath.Base(full)).Option("missingkey=zero").ParseFiles(files...)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// Разные шаблоны в наборе — исполняем именно тот, что запросили.
	if err := t.ExecuteTemplate(&buf, filepath.Base(full), data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
