package pdf

import (
	"github.com/pro-assistance-dev/sprob/helpers/templater"
	"github.com/unidoc/unidoc/pdf/creator"
)

type PDF struct {
	templater *templater.Templater
	// generator *wkhtmltopdf.PDFGenerator

	creator *creator.Creator
	ws      *mywriter
}

// func NewPDF(config config.Project) *PDF {
// 	pdfg, err := wkhtmltopdf.NewPDFGenerator()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return &PDF{
// 		ws:        &mywriter{},
// 		templater: templater.NewTemplater(config),
// 		// generator: pdfg,
// 	}
// }

// func (i *PDF) GeneratePDF(template string, data interface{}) ([]byte, error) {
// 	dataString := i.templater.Parse(template, data)
// 	gen, err := wkhtmltopdf.NewPDFGenerator()
// 	if err != nil {
// 		return nil, err
// 	}
// 	i.generator = gen
// 	i.writeNewPageFromString(dataString)
// 	i.setPageOptions()
// 	return i.createFile()
// }

func (i *PDF) MergeFilesToPDF(files IFiles) ([]byte, error) {
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

// func (i *PDF) setPageOptions() {
// 	i.generator.PageSize.Set(wkhtmltopdf.PageSizeA4)
// 	i.generator.Dpi.Set(300)
// }

// func (i *PDF) createFile() ([]byte, error) {
// 	err := i.generator.Create()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return i.generator.Bytes(), nil
// }

// func (i *PDF) writeNewPageFromString(data string) {
// 	i.generator.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(data))))
// }
