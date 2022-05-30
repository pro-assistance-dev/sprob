package pdfHelper

import (
	"errors"
	"fmt"
	"github.com/unidoc/unidoc/pdf/creator"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Mergeable interface {
	MergeTo(c *creator.Creator) error
}

type IFile interface {
	GetOriginalName() string
	GetFullPath() string
}

type IFiles []IFile

type source struct {
	path, mime string
}

// Initiate new source file from input argument
func (i *PDFHelper) newSource(file IFile) Mergeable {
	var inputSource Mergeable

	//if len(fileInputParts) > 1 {
	//	pages = parsePageNums(fileInputParts[1])
	//}
	inputSource = getMergeableFile(file)

	return inputSource
}

func getMergeableFile(file IFile) Mergeable {
	f, err := os.Open(file.GetFullPath())
	if err != nil {
		log.Fatal("Cannot read source file:", file.GetFullPath())
	}

	defer f.Close()

	ext := filepath.Ext(file.GetOriginalName())
	mime, err := getMimeType(f)
	if err != nil {
		log.Fatal("Error in getting mime type of file:", file.GetOriginalName())
	}

	sourceType, err := getFileType(mime, ext)
	if err != nil {
		log.Fatal("Error : %s (%s)", err.Error(), file.GetFullPath())
	}

	source := source{file.GetFullPath(), mime}

	var m Mergeable
	switch sourceType {
	case "image":
		m = ImgSource{source}
	case "pdf":
		m = PDFSource{source}
	}

	return m
}

func getFileType(mime, ext string) (string, error) {
	pdfExts := []string{".pdf", ".PDF"}
	imgExts := []string{".jpg", ".jpeg", ".gif", ".png", ".tiff", ".tif", ".JPG", ".JPEG", ".GIF", ".PNG", ".TIFF", ".TIF"}

	switch {
	case mime == "application/pdf":
		return "pdf", nil
	case mime[:6] == "image/":
		return "image", nil
	case mime == "application/octet-stream" && in_array(ext, pdfExts):
		return "pdf", nil
	case mime == "application/octet-stream" && in_array(ext, imgExts):
		return "image", nil
	}

	return "error", errors.New("File type not acceptable. ")
}

func parsePageNums(pagesInput string) []int {
	pages := []int{}

	for _, e := range strings.Split(pagesInput, ",") {
		pageNo, err := strconv.Atoi(strings.Trim(e, " \n"))
		if err != nil {
			fmt.Errorf("Invalid format! Example of a file input with page numbers: path/to/abc.pdf~1,2,3,5,6")
			os.Exit(1)
		}
		pages = append(pages, pageNo)
	}

	return pages
}
