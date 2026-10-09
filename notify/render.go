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

// renderFile — рендер шаблона-файла. Для HTML-шаблонов (`.gohtml`) к набору
// добавляются общие `_header.html`/`_footer.html` (если есть рядом с TEMPLATES_PATH);
// прочие (телеграм-текст и т.п.) рендерятся в одиночку.
// `root` — корень данных шаблона (обычно богатый объект события).
func renderFile(relPath string, root any) (string, error) {
	dir := strings.TrimSpace(os.Getenv("TEMPLATES_PATH"))
	if dir == "" {
		dir = "/app/templates"
	}
	full := filepath.Join(dir, relPath)
	files := []string{full}
	if strings.HasSuffix(relPath, ".gohtml") {
		for _, shared := range []string{"_header.html", "_footer.html"} {
			p := filepath.Join(dir, shared)
			if _, err := os.Stat(p); err == nil {
				files = append(files, p)
			}
		}
	}
	t, err := template.New(filepath.Base(full)).Option("missingkey=zero").ParseFiles(files...)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// Разные шаблоны в наборе — исполняем именно тот, что запросили.
	if err := t.ExecuteTemplate(&buf, filepath.Base(full), root); err != nil {
		return "", err
	}
	return buf.String(), nil
}
