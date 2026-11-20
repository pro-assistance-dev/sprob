package newsfeed

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Источник ВКонтакте: посты стены группы через `wall.get`.
//
// ⚠️ ПОЧЕМУ API, А НЕ HTML (замер sred 19.11.2026): API отдаёт ПОЛНЫЙ текст
// поста (самый длинный в группе — 4 КБ), а веб-версия режет на «показать
// полностью» и требует JS. Статьи ВК (`vk.ru/@slug`) API не отдаёт вовсе —
// они синхронизируются отдельно (sred/vkarticles), в этот источник не входят.
//
// ⚠️ `filter=owner` — только посты самой группы: без него в ленту попадают
// репосты чужих текстов, и на сайте это выглядит плагиатом (проверено в sred).
type VK struct {
	// OwnerID — id группы/сообщества: «123456» или «-123456» (минус не нужен).
	OwnerID string
	// Token — сервисный/групповой ключ с доступом к wall.get.
	Token string
	// Count — сколько постов брать (по умолчанию 50, максимум 100).
	Count int
	// Fetcher — HTTP-клиент (таймауты и повторы — общие с Telegram).
	Fetcher Fetch
}

var vkAPIBase = "https://api.vk.com/method/"

// vkResponse — обёртка ответа ВК (ошибки приходят в теле, а не кодом HTTP).
type vkResponse[T any] struct {
	Response T `json:"response"`
	Error    *struct {
		Code int    `json:"error_code"`
		Msg  string `json:"error_msg"`
	} `json:"error"`
}

type vkWall struct {
	Count int          `json:"count"`
	Items []vkWallPost `json:"items"`
}

type vkWallPost struct {
	ID        int    `json:"id"`
	OwnerID   int    `json:"owner_id"`
	Date      int64  `json:"date"`
	Text      string `json:"text"`
	IsPinned  int    `json:"is_pinned"`
	MarkedAs  int    `json:"marked_as_ads"`
	Copyright struct {
		Link string `json:"link"`
	} `json:"copyright"`
	Attachments []vkAttachment `json:"attachments"`
}

type vkAttachment struct {
	Type  string `json:"type"`
	Photo *struct {
		Text  string `json:"text"`
		Sizes []struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			Type   string `json:"type"`
		} `json:"sizes"`
	} `json:"photo"`
	Link *struct {
		URL     string `json:"url"`
		Title   string `json:"title"`
		Caption string `json:"caption"`
	} `json:"link"`
	Video *struct {
		ID        int    `json:"id"`
		OwnerID   int    `json:"owner_id"`
		Title     string `json:"title"`
		Player    string `json:"player"`
		AccessKey string `json:"access_key"`
	} `json:"video"`
}

// Posts — забирает посты стены.
func (v VK) Posts() ([]Post, error) {
	owner := strings.TrimPrefix(strings.TrimSpace(v.OwnerID), "-")
	if owner == "" {
		return nil, fmt.Errorf("vk: не задан owner id")
	}
	if v.Token == "" {
		return nil, fmt.Errorf("vk: не задан токен (VK_TOKEN)")
	}
	count := v.Count
	if count <= 0 || count > 100 {
		count = 50
	}

	q := url.Values{
		"owner_id": {"-" + owner},
		"count":    {strconv.Itoa(count)},
		// ⚠️ filter=owner: только посты группы. Без него в ленту попадают
		// репосты чужих текстов (проверено в sred — выглядело как плагиат).
		"filter":       {"owner"},
		"access_token": {v.Token},
		"v":            {"5.199"},
	}

	body, err := v.Fetcher.Get(vkAPIBase + "wall.get?" + q.Encode())
	if err != nil {
		return nil, fmt.Errorf("vk wall.get: %w", err)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("vk: пустой ответ")
	}
	var out vkResponse[vkWall]
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("vk: ответ не распознан: %.200s", body)
	}
	if out.Error != nil {
		// ⚠️ Типичная ошибка «access_token is invalid» — ключ от другого
		// типа приложения. Пишем код и текст, чтобы не гадать.
		return nil, fmt.Errorf("vk: %d %s", out.Error.Code, out.Error.Msg)
	}

	posts := make([]Post, 0, len(out.Response.Items))
	for _, raw := range out.Response.Items {
		// Реклама и закреплённые посты в ленту не идут: закреп висит сверху
		// постоянно и при каждом синке «обновлял» бы одно и то же место.
		if raw.MarkedAs == 1 || raw.IsPinned == 1 {
			continue
		}
		p := v.convert(raw)
		if !hasContent(p) || SkipText(p.Text()) {
			continue
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (v VK) convert(raw vkWallPost) Post {
	p := Post{
		ID:          strconv.Itoa(raw.ID),
		URL:         fmt.Sprintf("https://vk.com/wall-%d_%d", absInt(raw.OwnerID), raw.ID),
		PublishedAt: time.Unix(raw.Date, 0).UTC(),
	}
	text := CleanSpaces(CleanArtifacts(raw.Text))
	if text != "" {
		p.Blocks = append(p.Blocks, Block{Kind: BlockText, Text: text})
	}

	var photos []string
	for _, a := range raw.Attachments {
		switch a.Type {
		case "photo":
			if a.Photo == nil {
				continue
			}
			if u := bestPhoto(a.Photo.Sizes); u != "" {
				photos = append(photos, u)
			}
		case "link":
			if a.Link != nil && a.Link.URL != "" {
				label := a.Link.Title
				if label == "" {
					label = a.Link.URL
				}
				p.Blocks = append(p.Blocks, Block{
					Kind:  BlockLink,
					Text:  CleanSpaces(label),
					Links: []Link{{Text: CleanSpaces(label), Href: a.Link.URL}},
				})
			}
		case "video":
			if a.Video == nil {
				continue
			}
			// VK отдаёт готовый player-URL для iframe.
			p.Blocks = append(p.Blocks, Block{
				Kind:     BlockVideo,
				Text:     CleanSpaces(a.Video.Title),
				Embed:    a.Video.Player,
				VideoURL: fmt.Sprintf("https://vk.com/video-%d_%d", absInt(a.Video.OwnerID), a.Video.ID),
			})
		}
	}
	if len(photos) > 0 {
		p.Blocks = append(p.Blocks, Block{Kind: BlockImage, Images: photos})
	}
	return Post{ID: p.ID, Blocks: GroupGallery(p.Blocks), PublishedAt: p.PublishedAt, URL: p.URL}
}

// bestPhoto — самая крупная из предложенных размеров (VK даёт лестницу
// s/m/x/y/z/w). Берём по площади: у части постов нет «w», но есть «z».
func bestPhoto(sizes []struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   string `json:"type"`
}) string {
	best, bestArea := "", 0
	for _, s := range sizes {
		if s.URL == "" {
			continue
		}
		if area := s.Width * s.Height; area >= bestArea {
			best, bestArea = s.URL, area
		}
	}
	return best
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// setVKBase — подмена адреса API в тестах (в проде всегда api.vk.com).
func setVKBase(u string) { vkAPIBase = u }
