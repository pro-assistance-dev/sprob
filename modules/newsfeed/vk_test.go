package newsfeed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ⚠️ Форма ответа ВК (усечённая, но с реальной структурой): текст, фото
// лестницей размеров, ссылка, видео с player-URL, реклама и закреп.
const vkJSON = `{"response":{"count":4,"items":[
 {"id":10,"owner_id":-123,"date":1700000000,"text":"Первый пост с фото",
 "attachments":[{"type":"photo","photo":{"sizes":[
 {"url":"https://vk/small.jpg","width":75,"height":75,"type":"s"},
 {"url":"https://vk/big.jpg","width":1080,"height":1080,"type":"w"}]}}]},
 {"id":11,"owner_id":-123,"date":1700000100,"text":"Пост со ссылкой и видео",
 "attachments":[
 {"type":"link","link":{"url":"https://example.com/a","title":"Подробнее"}},
 {"type":"video","video":{"id":5,"owner_id":-123,"title":"Ролик","player":"https://vk.com/video_ext.php?oid=-123&id=5"}}]},
 {"id":12,"owner_id":-123,"date":1700000200,"text":"Реклама","marked_as_ads":1},
 {"id":13,"owner_id":-123,"date":1700000300,"text":"Закреплённый пост","is_pinned":1}
]}}`

func vkServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что запрос идёт с filter=owner и токеном
		if !strings.Contains(r.URL.RawQuery, "filter=owner") {
			t.Errorf("нет filter=owner в запросе: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "access_token=tok") {
			t.Errorf("нет токена в запросе")
		}
		_, _ = w.Write([]byte(vkJSON))
	}))
}

// Реализуем подмену базы API — иначе тест полез бы в настоящий vk.com.
func TestVK_Posts(t *testing.T) {
	srv := vkServer(t)
	defer srv.Close()

	// временно подменяем базу (в проде — api.vk.com)
	old := vkAPIBase
	defer func() { setVKBase(old) }()
	setVKBase(srv.URL + "/method/")

	posts, err := VK{OwnerID: "-123", Token: "tok"}.Posts()
	if err != nil {
		t.Fatalf("Posts: %v", err)
	}
	// реклама и закреп отфильтрованы
	if len(posts) != 2 {
		t.Fatalf("постов %d, ожидали 2 (без рекламы и закрепа): %+v", len(posts), posts)
	}

	// 1) фото: берём САМЫЙ КРУПНЫЙ размер, а не первый
	p1 := posts[0]
	imgs := p1.Images()
	if len(imgs) != 1 || imgs[0] != "https://vk/big.jpg" {
		t.Errorf("фото: %v, ожидали big.jpg (1080×1080)", imgs)
	}
	if p1.Blocks[0].Kind != BlockText || !strings.Contains(p1.Text(), "Первый пост") {
		t.Errorf("текст поста потерялся: %+v", p1.Blocks)
	}

	// 2) ссылка и видео — отдельными блоками
	p2 := posts[1]
	var hasLink, hasVideo bool
	for _, b := range p2.Blocks {
		if b.Kind == BlockLink && len(b.Links) == 1 && b.Links[0].Href == "https://example.com/a" {
			hasLink = true
		}
		if b.Kind == BlockVideo && strings.Contains(b.Embed, "video_ext.php") {
			hasVideo = true
		}
	}
	if !hasLink {
		t.Errorf("ссылка потерялась: %+v", p2.Blocks)
	}
	if !hasVideo {
		t.Errorf("видео потерялось: %+v", p2.Blocks)
	}

	// дата — из Unix-времени поста, не «сейчас»
	if p1.PublishedAt.Unix() != 1700000000 {
		t.Errorf("дата %v, ожидали Unix 1700000000", p1.PublishedAt)
	}
}

func TestVK_NotConfigured(t *testing.T) {
	if _, err := (VK{OwnerID: "1"}).Posts(); err == nil {
		t.Error("без токена должна быть ошибка")
	}
	if _, err := (VK{Token: "x"}).Posts(); err == nil {
		t.Error("без owner id должна быть ошибка")
	}
}
