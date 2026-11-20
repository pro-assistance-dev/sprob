package newsfeed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ⚠️ СКВОЗНОЙ тест на реальный текст поста (20.11.2026): ссылка на альбом,
// упоминание и хвост из хештегов — всё вместе, как в жизни.
func TestVK_RealPost_AllTextForms(t *testing.T) {
	const post = `{"response":{"count":1,"items":[
	 {"id":6755,"owner_id":-204353467,"date":1732000000,
	 "text":"ФОТКИ! СМОТРИТЕ В АЛЬБОМЕ! https://vk.ru/album-119823711_311911110 Спасибо [club119823711|АртФаина] за кадры!\n\n#средневековка2026 #фаершоу #огонь"}]}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(post))
	}))
	defer srv.Close()
	old := vkAPIBase
	defer setVKBase(old)
	setVKBase(srv.URL + "/method/")

	posts, err := VK{OwnerID: "204353467", Token: "tok"}.Posts()
	if err != nil {
		t.Fatalf("Posts: %v", err)
	}
	p := posts[0]
	text := p.Text()

	// 1. хештегов нет
	if strings.Contains(text, "#") {
		t.Errorf("хештеги остались: %q", text)
	}
	// 2. разметки упоминания нет, подпись осталась
	if strings.Contains(text, "[club") || !strings.Contains(text, "АртФаина") {
		t.Errorf("упоминание: %q", text)
	}
	// 3. обе ссылки на месте
	links := p.Links()
	var album, club bool
	for _, l := range links {
		if strings.HasPrefix(l.Href, "https://vk.ru/album-119823711_311911110") {
			album = true
		}
		if l.Href == "https://vk.com/club119823711" {
			club = true
		}
	}
	if !album {
		t.Errorf("ссылка на альбом потерялась: %+v", links)
	}
	if !club {
		t.Errorf("ссылка на сообщество потерялась: %+v", links)
	}
}
