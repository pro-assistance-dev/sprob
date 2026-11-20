package newsfeed

import (
	"html"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ParseHTML — HTML поста → блоки.
//
// ⚠️ ЗАЧЕМ (20.11.2026). Раньше текст поста «просто снимали с тегов»: от
// `<a href="https://music.yandex.ru/...">Яндекс Музыка</a>` оставалась подпись
// «Яндекс Музыка» — на сайте она выглядела ссылкой, но нажать её было нельзя.
// Проверено на живой ленте: «Слушать Яндекс Музыка | VK Музыка» — мёртвый текст.
// Теперь ссылки, картинки и цитаты выделяются в структуру (Link/Block).
//
// Почему `golang.org/x/net/html`, а не регулярки: `<a><b>текст</b></a>` на
// регулярках рвётся, а порядок атрибутов не фиксирован (darbore уже попадался
// на этом с картинками t.me). Пакет уже есть в go.mod sprob.
//
// Правила (сознательные):
// - блочные теги (p, div) → границы абзацев, <br> → перевод строки;
// - <b>/<i> прозрачны: форматирование теряется, текст остаётся;
// - <a> → Link внутри блока; ссылка «островом» (без текста вокруг) →
// отдельный BlockLink, чтобы клиент нарисовал кнопку;
// - <img> → BlockImage, по порядку;
// - <blockquote> → BlockQuote, <ul>/<ol> → BlockList.
func ParseHTML(raw string) []Block {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// ⚠️ Контекст разбора — BODY, и обязательно с DataAtom: x/net/html сверяет
	// `Data` и `DataAtom` и на рассогласовании отдаёт
	// `html: inconsistent Node: DataAtom="", Data="body"` → ноль узлов.
	// Без атома тест 20.11 показал пустой пост (ссылки и картинки пропадали).
	ctx := &xhtml.Node{Type: xhtml.ElementNode, DataAtom: atom.Body, Data: "body"}
	nodes, err := xhtml.ParseFragment(strings.NewReader(raw), ctx)
	if err != nil {
		// HTML битый — не теряем пост: отдаём текстом как есть.
		if t := strings.TrimSpace(stripTags(raw)); t != "" {
			return []Block{{Kind: BlockText, Text: t}}
		}
		return nil
	}

	b := &blockBuilder{}
	for _, n := range nodes {
		b.walk(n)
	}
	return b.finish()
}

func (b *blockBuilder) walk(n *xhtml.Node) {
	switch n.Type {
	case xhtml.TextNode:
		b.text.WriteString(n.Data)
		return
	case xhtml.DocumentNode:
		b.walkChildren(n)
		return
	case xhtml.ElementNode:
	default:
		return
	}

	switch n.Data {
	case "script", "style", "iframe":
		return
	case "img":
		b.flushText()
		if src := attr(n, "src"); src != "" {
			b.blocks = append(b.blocks, Block{Kind: BlockImage, Images: []string{absURL(src)}})
		}
		return
	case "br":
		b.text.WriteString("\n")
		return
	case "p", "div", "section", "article":
		b.flushText()
		b.walkChildren(n)
		b.flushText()
		return
	case "blockquote":
		b.flushText()
		b.inQuote++
		b.walkChildren(n)
		b.inQuote--
		b.flushText()
		return
	case "ul", "ol":
		b.flushText()
		b.inList++
		b.walkChildren(n)
		b.inList--
		b.flushList()
		return
	case "li":
		b.flushText()
		b.walkChildren(n)
		if item := strings.TrimSpace(b.takeText()); item != "" {
			b.list = append(b.list, item)
		}
		return
	case "a":
		b.onAnchor(n)
		return
	}
	// b, i, u, span, strong… — прозрачные
	b.walkChildren(n)
}

func (b *blockBuilder) walkChildren(n *xhtml.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.walk(c)
	}
}

// onAnchor: ссылка внутри фразы или отдельная «кнопка».
func (b *blockBuilder) onAnchor(n *xhtml.Node) {
	href := attr(n, "href")
	label := strings.TrimSpace(textOf(n))
	if href == "" || label == "" {
		if label != "" {
			b.text.WriteString(label)
		}
		return
	}
	href = absURL(href)

	// ⚠️ Хештеги Telegram/VK размечает как ссылки на поиск
	// (`https://t.me/<канал>/<id>?q=%23нуар`). Нам они не нужны и только
	// засоряют карточку: подпись остаётся текстом, адрес выкидываем.
	// Проверено на живой ленте 20.11: 3 ссылки-хештега в каждом посте.
	if isHashtagAnchor(label, href) {
		b.text.WriteString(label)
		return
	}

	// Вокруг ссылки ещё нет текста → самостоятельная кнопка (так Telegram
	// подаёт «Яндекс Музыка | VK Музыка» отдельными строками).
	if strings.TrimSpace(b.text.String()) == "" {
		b.flushText()
		b.blocks = append(b.blocks, Block{
			Kind:  BlockLink,
			Text:  label,
			Links: []Link{{Text: label, Href: href}},
		})
		return
	}
	b.text.WriteString(label)
	b.links = append(b.links, Link{Text: label, Href: href})
}

// blockBuilder накапливает блоки по мере обхода дерева.
type blockBuilder struct {
	blocks  []Block
	text    strings.Builder
	links   []Link
	inQuote int
	inList  int
	list    []string
}

func (b *blockBuilder) takeText() string {
	t := b.text.String()
	b.text.Reset()
	return cleanSpaces(t)
}

func (b *blockBuilder) flushText() {
	t := b.takeText()
	if t == "" {
		b.links = nil
		return
	}
	kind := BlockText
	if b.inQuote > 0 {
		kind = BlockQuote
	}
	b.blocks = append(b.blocks, Block{Kind: kind, Text: t, Links: b.links})
	b.links = nil
}

func (b *blockBuilder) flushList() {
	if len(b.list) == 0 {
		return
	}
	b.blocks = append(b.blocks, Block{Kind: BlockList, Text: strings.Join(b.list, "\n")})
	b.list = nil
}

func (b *blockBuilder) finish() []Block {
	b.flushText()
	b.flushList()
	return b.blocks
}

// stripTags — страховочный «просто текст» для битого HTML.
func stripTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			sb.WriteRune(r)
		}
	}
	return html.UnescapeString(sb.String())
}

func attr(n *xhtml.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func textOf(n *xhtml.Node) string {
	var sb strings.Builder
	var rec func(*xhtml.Node)
	rec = func(x *xhtml.Node) {
		if x.Type == xhtml.TextNode {
			sb.WriteString(x.Data)
			return
		}
		if x.Type == xhtml.ElementNode && (x.Data == "script" || x.Data == "style") {
			return
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	rec(n)
	// ⚠️ NBSP → пробел и здесь: подпись ссылки («Д.\u00a0Зубарев») приходит
	// из текстового узла и минует cleanSpaces. Телеграм ставит NBSP в инициалы.
	return strings.ReplaceAll(sb.String(), "\u00a0", " ")
}

// cleanSpaces — схлопывает пробелы внутри строк, не больше одной пустой
// строки подряд, без пустых строк по краям.
//
// ⚠️ Переводы строк ЗНАЧИМЫ (это абзацы поста), поэтому режем только
// горизонтальные пробелы и «лесенку» из пустых строк.
func cleanSpaces(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, ln := range lines {
		// ⚠️ NBSP (\u00a0) → обычный пробел: Telegram вставляет его в
		// инициалы («Д.\u00a0Зубарев»), и в JSON/поиске это ломает сравнение.
		ln = strings.ReplaceAll(ln, "\u00a0", " ")
		ln = strings.Join(strings.Fields(ln), " ")
		if ln == "" {
			blank++
			if blank > 1 || len(out) == 0 {
				continue
			}
			out = append(out, "")
			continue
		}
		blank = 0
		out = append(out, ln)
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n")
}

// CleanSpaces — экспорт для проектов, которые чистят текст сами.
func CleanSpaces(s string) string { return cleanSpaces(s) }

// absURL — протокол-относительные адреса (`//cdn…`) → https. Telegram и ВК
// отдают такие в картинках; без этого `<img src="//cdn…">` на сайте не грузится.
func absURL(u string) string {
	u = strings.TrimSpace(u)
	if strings.HasPrefix(u, "//") {
		return "https:" + u
	}
	return u
}

// isHashtagAnchor — ссылка-хештег (поиск по тегу в Telegram/VK).
func isHashtagAnchor(label, href string) bool {
	if !strings.HasPrefix(label, "#") {
		return false
	}
	// поиск по тегу: t.me/<канал>/<id>?q=%23…, vk.ru/search/statuses?q=%23…
	return strings.Contains(href, "?q=%23") || strings.Contains(href, "?q=#") ||
		strings.Contains(href, "/search/")
}
