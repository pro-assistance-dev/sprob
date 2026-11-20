package newsfeed

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Источник Telegram: веб-превью `t.me/s/<канал>` (основной, если доступно)
// и RSSHub-инстанс (работает из РФ — сам ходит в Telegram).
//
// ⚠️ Почему два и в каком порядке (20.11.2026, замеры с прод-сервера):
// - `t.me` из РФ заблокирован (с сервера: connection timed out);
// - RSSHub работает (0.7–1.2 с на полный документ), но бывает отдаёт
// обрезанный документ (~16 КБ из 55 КБ, без `</rss>`).
// Поэтому по умолчанию — RSSHub, а обрезанный документ = повод перейти на t.me.

// Telegram — источник ленты Telegram-канала.
type Telegram struct {
	// Channel — username канала без «@».
	Channel string
	// RSSURL — адрес RSSHub-инстанса для этого канала.
	RSSURL string
	// Prefer — "rsshub" (по умолчанию) или "tme".
	Prefer string
	// Fetcher — HTTP-клиент; нулевое значение = настройки по умолчанию.
	Fetcher Fetch
}

// Posts — забирает последние посты канала.
func (t Telegram) Posts() ([]Post, error) {
	prefer := strings.ToLower(strings.TrimSpace(t.Prefer))
	if prefer == "" {
		prefer = "rsshub"
	}
	first, second := t.fromRSS, t.fromTme
	if prefer == "tme" {
		first, second = t.fromTme, t.fromRSS
	}

	posts, errFirst := first()
	if errFirst == nil && len(posts) > 0 {
		return posts, nil
	}
	posts, errSecond := second()
	if errSecond == nil && len(posts) > 0 {
		return posts, nil
	}
	if errFirst != nil {
		return nil, fmt.Errorf("оба источника недоступны: %v / %v", errFirst, errSecond)
	}
	return nil, fmt.Errorf("источники вернули пустую ленту: %v", errSecond)
}

// ── RSSHub ───────────────────────────────────────────────────────────────

type rssDoc struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (t Telegram) fromRSS() ([]Post, error) {
	if strings.TrimSpace(t.RSSURL) == "" {
		return nil, fmt.Errorf("rsshub: адрес не задан (TELEGRAM_NEWS_RSS_URL)")
	}
	// ⚠️ Обрыв документа (darbore, 29.10): RSSHub рвётся на медиа-вставке и
	// отдаёт кусок без `</rss>`. `xml.Unmarshal` скажет `unexpected EOF` —
	// это выглядит как «RSSHub недоступен», хотя причина другая (битый ответ).
	// ⚠️ Плюс нестабильность (20.11): один запрос — 0 байт, следующий — полный
	// документ. Поэтому GetComplete: проверка `</rss>` + повтор.
	body, err := t.Fetcher.GetComplete(t.RSSURL, "rss")
	if err != nil {
		return nil, fmt.Errorf("rsshub: %w", err)
	}
	var doc rssDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("rsshub xml: %w", err)
	}

	out := make([]Post, 0, len(doc.Channel.Items))
	for _, it := range doc.Channel.Items {
		id := postIDFromLink(it.Link)
		if id == "" {
			continue
		}
		blocks := ParseHTML(CleanArtifacts(it.Description))
		p := Post{
			ID:          id,
			Title:       cleanTitle(it.Title),
			Blocks:      GroupGallery(blocks),
			PublishedAt: parsePubDate(it.PubDate),
			URL:         strings.TrimSpace(it.Link),
		}
		if p.PublishedAt.IsZero() {
			p.PublishedAt = time.Now().UTC()
		}
		if !hasContent(p) || SkipText(p.Text()) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// cleanTitle — заголовок поста как заголовок, а не как строка «Канал —
// Telegram Channel» (RSSHub дублирует её в каждом элементе).
func cleanTitle(t string) string {
	t = strings.TrimSpace(t)
	if strings.HasSuffix(strings.ToLower(t), "— telegram channel") ||
		strings.HasSuffix(strings.ToLower(t), "- telegram channel") {
		return ""
	}
	return CleanSpaces(t)
}

// postIDFromLink — номер поста из ссылки `https://t.me/<канал>/<id>`.
// Возвращает "" для ссылок без числового id (закреп, опрос и т. п.).
func postIDFromLink(link string) string {
	link = strings.TrimSpace(link)
	idx := strings.LastIndex(link, "/")
	if idx < 0 || idx == len(link)-1 {
		return ""
	}
	id := link[idx+1:]
	if _, err := strconv.ParseInt(id, 10, 64); err != nil {
		return ""
	}
	return id
}

// parsePubDate — дата публикации в форматах, которые отдают RSSHub и t.me.
// ⚠️ Раньше (ferma) дата не использовалась вовсе: при первом импорте вся
// лента получала одну дату «сейчас», и хронология врала (20.11).
func parsePubDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// hasContent — есть ли в посте хоть что-то для показа.
func hasContent(p Post) bool {
	return p.Text() != "" || len(p.Images()) > 0 || len(p.Links()) > 0
}

// ── t.me (веб-превью) ────────────────────────────────────────────────────

var (
	postWrapRe = regexp.MustCompile(`data-post="([^"/]+)/(\d+)"`)
	textRe     = regexp.MustCompile(`(?s)<div class="tgme_widget_message_text[^"]*"[^>]*>(.*?)</div>`)
	// ⚠️ Порядок атрибутов в t.me НЕ фиксирован: `href` может идти до `style`
	// (darbore ловил это 25.10: регулярка с `class` перед `style` не находила
	// картинок, и все посты шли без фото). Ловим тег целиком.
	photoARe   = regexp.MustCompile(`(?s)<a[^>]*class="[^"]*tgme_widget_message_photo_wrap[^"]*"[^>]*>`)
	photoURLRe = regexp.MustCompile(`url\('([^']+)'\)`)
	timeRe     = regexp.MustCompile(`<time[^>]*datetime="([^"]+)"`)
)

func (t Telegram) fromTme() ([]Post, error) {
	ch := strings.TrimPrefix(strings.TrimSpace(t.Channel), "@")
	if ch == "" {
		return nil, fmt.Errorf("t.me: не задан канал")
	}
	body, err := t.Fetcher.Get("https://t.me/s/" + ch)
	if err != nil {
		return nil, fmt.Errorf("t.me: %w", err)
	}
	s := string(body)
	blocksHTML := strings.Split(s, "tgme_widget_message_wrap")
	out := make([]Post, 0, len(blocksHTML)-1)
	for _, b := range blocksHTML[1:] {
		m := postWrapRe.FindStringSubmatch(b)
		if len(m) < 3 {
			continue
		}
		p := Post{
			ID:  m[2],
			URL: "https://t.me/" + ch + "/" + m[2],
		}
		if tm := textRe.FindStringSubmatch(b); len(tm) > 1 {
			p.Blocks = ParseHTML(CleanArtifacts(tm[1]))
		}
		var photos []string
		for _, tag := range photoARe.FindAllString(b, -1) {
			if pm := photoURLRe.FindStringSubmatch(tag); len(pm) > 1 {
				if u := absURL(pm[1]); u != "" {
					photos = append(photos, u)
				}
			}
		}
		if len(photos) > 0 {
			p.Blocks = append(p.Blocks, Block{Kind: BlockImage, Images: photos})
		}
		if tm := timeRe.FindStringSubmatch(b); len(tm) > 1 {
			if ts, err := time.Parse(time.RFC3339, tm[1]); err == nil {
				p.PublishedAt = ts.UTC()
			}
		}
		if p.PublishedAt.IsZero() {
			p.PublishedAt = time.Now().UTC()
		}
		p.Blocks = GroupGallery(p.Blocks)
		if !hasContent(p) || SkipText(p.Text()) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}
