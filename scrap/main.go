package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
)

const site = "https://rdkb.ru"

func main() {
	Scrap()
}

func Scrap() {
	// menus := []string{"about"}
	// for _, m := range menus {
	// fmt.Println(m)
	// page, _ := url.JoinPath(site, m)
	// fileName := fmt.Sprintf("%s.json", m)
	geziyor.NewGeziyor(&geziyor.Options{
		StartURLs: []string{"https://rdkb.ru/about/ustav.php"},
		ParseFunc: parse,
		// Exporters: []export.Exporter{&export.JSON{FileName: fileName, EscapeHTML: true}},
	}).Start()
	// }
}

func parse(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("div.body-wrapper").Each(func(_ int, s *goquery.Selection) {
		// content, _ := s.Find("div.container.container-main > div > div.col.col-9 > div").Text()
		contentS := s.Find("div.container.container-main > div > div.col.col-9 > div")
		// title, _ := s.Find("div.container.container-h1.container-white > div > div.col.col-12.mb20 > h1").Html()
		contentS.Find("a").Each(func(_ int, link *goquery.Selection) {
			fileName, _ := link.Attr("href")
			// text := link.Text()
			// fmt.Println(href, text)
			download(fileName)
		})

		// g.Exports <- map[string]interface{}{
		// 	"title":   title,
		// 	"content": contentS.Text(),
		// }
	})

	// r.HTMLDoc.Find("li.bx_hma_one_lvl").Each(func(_ int, s *goquery.Selection) {
	// 	if href, ok := s.Find("a").Attr("href"); ok {
	// 		g.Get(r.JoinURL(href), parse)
	// 	}
	// })
}

func download(link string) {
	out, err := os.Create(strings.ReplaceAll(link, "/files/", ""))
	fmt.Println(err)
	defer func() {
		err = out.Close()
		fmt.Println(err)
	}()

	fileLink, err := url.JoinPath(site, link)
	fmt.Println(err)
	resp, err := http.Get(fileLink)
	fmt.Println(err)
	defer func() {
		err = resp.Body.Close()
		fmt.Println(err)
	}()

	_, _ = io.Copy(out, resp.Body)
}
