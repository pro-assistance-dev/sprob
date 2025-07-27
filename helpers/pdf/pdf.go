package pdf

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/unidoc/unidoc/pdf/creator"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

type PDFSource struct { //nolint:golint
	source
}

func (s PDFSource) MergeTo(c *creator.Creator) error {
	f, _ := os.Open(s.path)
	defer func() {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	return addPdfPages(f, c)
}

func getReader(rs io.ReadSeeker) (*pdf.PdfReader, error) {
	pdfReader, err := pdf.NewPdfReader(rs)
	if err != nil {
		return nil, err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}

	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, err
		}
		if !auth {
			return nil, errors.New("cannot merge encrypted, password protected document")
		}
	}

	return pdfReader, nil
}

func addPdfPages(file *os.File, c *creator.Creator) error {
	pdfReader, err := getReader(file)
	if err != nil {
		return err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}
		fmt.Println(page)
		if err = c.AddPage(page); err != nil {
			return err
		}
	}

	return nil
}
