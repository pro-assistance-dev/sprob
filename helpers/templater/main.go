package templater

import (
	"bytes"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lukasjarosch/go-docx"

	"pro-assister/config"
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

	// err := ioutil.WriteFile("./application-generate.html", []byte(buf.String()), 0644)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	return buf.String(), nil
}

func (i *Templater) ReplaceDoc(dataForReplacing map[string]interface{}, templatePath string) ([]byte, error) {
	templatePath = path.Join(i.templatesPath, templatePath)
	doc, err := docx.Open(templatePath)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = doc.ReplaceAll(dataForReplacing)
	if err != nil {
		return nil, err
	}
	err = doc.Write(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
