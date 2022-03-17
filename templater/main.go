package templater

import (
	"bytes"
	"fmt"
	"github.com/pro-assistance/pro-assister/config"
	"log"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type Templater struct {
	templatesPath string
}

func NewTemplater(config config.Config) *Templater {
	return &Templater{templatesPath: config.TemplatesPath}
}

func (i *Templater) Parse(templateName string, data interface{}) string {
	var buf strings.Builder
	templateName = fmt.Sprintf("%s.gohtml", filepath.Join(i.templatesPath, templateName))
	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		log.Fatal(err)
	}
	_ = tmpl.Execute(&buf, data)
	strTmpl := buf.String()
	return strTmpl
}

// ParseTemplate func
func (i *Templater) ParseTemplate(data interface{}, templates ...string) (string, error) {
	templatesForParse := []string{}
	templates = append(templates, "_footer.html", "_header.html") //добавляем хэдер и футер к шаблонам
	for _, t := range templates {
		templatesForParse = append(templatesForParse, path.Join(i.templatesPath, t)) //к каждому шаблону приблавляем путь
	}
	t := template.Must(template.ParseFiles(templatesForParse...))
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
