package writers

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

// XlsxSheet — одна страница книги Excel.
type XlsxSheet struct {
	Name   string
	Header []string
	Rows   [][]interface{}
}

// XlsxSheets формирует многостраничную книгу Excel: по листу на раздел
// (ответы + агрегаты опроса). Шапка стилизуется, ширина колонок — по содержимому.
// Шаблон выверен в map/hr (`helpers/writers/export.go`); при следующем
// переиспользовании — вынести в sprob как общий helper.
func XlsxSheets(sheets []XlsxSheet) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	first := true
	for si, sh := range sheets {
		if len(sh.Rows) == 0 && len(sh.Header) == 0 {
			continue
		}
		sheet := "Sheet1"
		if first {
			if sh.Name != "" {
				if err := f.SetSheetName("Sheet1", sh.Name); err != nil {
					return nil, err
				}
				sheet = sh.Name
			}
			first = false
		} else {
			name := sh.Name
			if name == "" {
				name = fmt.Sprintf("Лист%d", si+1)
			}
			if _, err := f.NewSheet(name); err != nil {
				return nil, err
			}
			sheet = name
		}

		if err := writeSheet(f, sheet, sh, headerStyle); err != nil {
			return nil, err
		}
	}

	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func writeSheet(f *excelize.File, sheet string, sh XlsxSheet, headerStyle int) error {
	for i, h := range sh.Header {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
	}
	if len(sh.Header) > 0 {
		end, _ := excelize.CoordinatesToCellName(len(sh.Header), 1)
		if err := f.SetCellStyle(sheet, "A1", end, headerStyle); err != nil {
			return err
		}
	}

	widths := make([]int, len(sh.Header))
	for i, h := range sh.Header {
		widths[i] = utf8.RuneCountInString(h)
	}
	for r, row := range sh.Rows {
		for c, v := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+2)
			if err != nil {
				return err
			}
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				return err
			}
			if c < len(widths) {
				if w := utf8.RuneCountInString(cellString(v)); w > widths[c] {
					widths[c] = w
				}
			}
		}
	}
	for i, w := range widths {
		if w < 8 {
			w = 8
		}
		if w > 60 {
			w = 60
		}
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		if err := f.SetColWidth(sheet, col, col, float64(w+2)); err != nil {
			return err
		}
	}
	return f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})
}

func cellString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
