package qr

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strconv"
)

// Рендер QR: SVG и PNG с «тихой зоной» 4 модуля. Без внешних зависимостей —
// пакет не должен тянуть SDK ради картинки.

// Matrix — модульная матрица (true = тёмный модуль).
type Matrix [][]bool

func (m Matrix) Modules() int {
	return len(m)
}

// SVG — векторная картинка с «тихой зоной» в 4 модуля (требование стандарта).
func (m Matrix) SVG(size int) []byte {
	n := m.Modules()
	quiet := 4
	total := n + quiet*2
	var b bytes.Buffer
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`" height="`)
	b.WriteString(strconv.Itoa(size))
	b.WriteString(`" viewBox="0 0 `)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(` `)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(`" shape-rendering="crispEdges"><rect width="100%" height="100%" fill="#fff"/><path d="`)
	for y := range n {
		for x := range n {
			if !m[y][x] {
				continue
			}
			b.WriteString("M")
			b.WriteString(strconv.Itoa(x + quiet))
			b.WriteString(",")
			b.WriteString(strconv.Itoa(y + quiet))
			b.WriteString("h1v1h-1z")
		}
	}
	b.WriteString(`" fill="#000"/></svg>`)
	return b.Bytes()
}

// PNG — растровая картинка 1:1 по пикселю на модуль (чёткий край при печати).
func (m Matrix) PNG(size int) []byte {
	n := m.Modules()
	quiet := 4
	total := n + quiet*2
	scale := size / total
	if scale < 1 {
		scale = 1
	}
	img := image.NewPaletted(image.Rect(0, 0, total*scale, total*scale),
		color.Palette{color.White, color.Black})
	for y := range n {
		for x := range n {
			if !m[y][x] {
				continue
			}
			for dy := range scale {
				for dx := range scale {
					px := (x + quiet) * scale
					py := (y + quiet) * scale
					img.SetColorIndex(px+dx, py+dy, 1)
				}
			}
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}
