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
}

func NewPDFHelper(config config.Config) *PDFHelper {
	return &PDFHelper{
		templater.NewTemplater(config),
	}
}

func (i *PDFHelper) GeneratePDF(template string, data interface{}) ([]byte, error) {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}
	dataString := i.templater.Parse(template, data)

	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(dataString))))
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}
	return pdfg.Bytes(), nil
}
