package pdfHelper

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/pro-assistance/pro-assister/config"
	"github.com/pro-assistance/pro-assister/templater"
	"github.com/unidoc/unidoc/pdf/creator"
	"log"
)

type PDFHelper struct {
	templater *templater.Templater
	generator *wkhtmltopdf.PDFGenerator
	reader    *wkhtmltopdf.PageReader
	creator   *creator.Creator
	ws        *mywriter
}

func NewPDFHelper(config config.Config) *PDFHelper {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Fatal(err)
	}
	return &PDFHelper{
		ws:        &mywriter{},
		templater: templater.NewTemplater(config),
		generator: pdfg,
	}
}

func (i *PDFHelper) GeneratePDF(template string, data interface{}) ([]byte, error) {
	dataString := i.templater.Parse(template, data)
	i.writeNewPageFromString(dataString)
	i.setPageOptions()
	return i.createFile()
}

func (i *PDFHelper) MergeFilesToPDF(files IFiles) ([]byte, error) {
	i.creator = creator.New()
	for _, file := range files {
		err := i.newSource(file).MergeTo(i.creator)
		if err != nil {
			return nil, err
		}
	}
	err := i.creator.Write(i.ws)
	if err != nil {
		return nil, err
	}
	return i.ws.buf, nil
}

func (i *PDFHelper) setPageOptions() {
	i.generator.PageSize.Set(wkhtmltopdf.PageSizeA4)
	i.generator.Dpi.Set(300)
}

func (i *PDFHelper) createFile() ([]byte, error) {
	err := i.generator.Create()
	if err != nil {
		return nil, err
	}
	return i.generator.Bytes(), nil
}

func (i *PDFHelper) writeNewPageFromString(data string) {
	i.generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(data))))
}

func (i *PDFHelper) writeNewPageFromBytes(data []byte) {
	i.generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(data)))
}
