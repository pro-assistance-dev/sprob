package pdfHelper

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/templater"
	"log"
)

type PDFHelper struct {
	templater *templater.Templater
	Generator *wkhtmltopdf.PDFGenerator
	Reader    *wkhtmltopdf.PageReader
}

func NewPDFHelper(config config.Config) *PDFHelper {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}
	return &PDFHelper{
		templater: templater.NewTemplater(config),
		Generator: pdfg,
	}
}

func (i *PDFHelper) GeneratePDF(template string, data interface{}) ([]byte, error) {
	dataString := i.templater.Parse(template, data)
	i.writeNewPageFromString(dataString)
	i.setPageOptions()
	return i.createFile()
}

func (i *PDFHelper) MergeFilesToPDF(files [][]byte) ([]byte, error) {
	i.setPageOptions()
	for _, file := range files {
		i.writeNewPageFromBytes(file)
	}
	return i.createFile()
}

func (i *PDFHelper) setPageOptions() {
	i.Generator.PageSize.Set(wkhtmltopdf.PageSizeA4)
	i.Generator.Dpi.Set(300)
}

func (i *PDFHelper) createFile() ([]byte, error) {
	err := i.Generator.Create()
	if err != nil {
		return nil, err
	}
	return i.Generator.Bytes(), nil
}

func (i *PDFHelper) writeNewPageFromString(data string) {
	i.Generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(data))))
}

func (i *PDFHelper) writeNewPageFromBytes(data []byte) {
	i.Generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(data)))
}
