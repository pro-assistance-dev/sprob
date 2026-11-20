package newsfeed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ⚠️ Реальная структура поста vk-2274 (20.11.2026): СЕМЬ фото-вложений.
// Раньше в блок попадало ОДНО (лучшее) — галерею показать было нечем.
const vkMultiPhoto = `{"response":{"count":1,"items":[
 {"id":2274,"owner_id":-204353467,"date":1732000000,"text":"Заснеженная Средневековка",
 "attachments":[
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p1_s.jpg","width":75,"height":75},
 {"url":"https://sun/p1_w.jpg","width":2560,"height":1182}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p2_s.jpg","width":75,"height":75},
 {"url":"https://sun/p2_w.jpg","width":2560,"height":1440}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p3_w.jpg","width":1080,"height":1080}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p4_w.jpg","width":1280,"height":720}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p5_w.jpg","width":1080,"height":1350}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p6_w.jpg","width":1080,"height":608}]}},
 {"type":"photo","photo":{"sizes":[
 {"url":"https://sun/p7_w.jpg","width":1080,"height":608}]}}
 ]}
]}}`

func TestVK_CollectsAllPhotos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(vkMultiPhoto))
	}))
	defer srv.Close()

	old := vkAPIBase
	defer setVKBase(old)
	setVKBase(srv.URL + "/method/")

	posts, err := VK{OwnerID: "204353467", Token: "tok"}.Posts()
	if err != nil {
		t.Fatalf("Posts: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("постов %d", len(posts))
	}
	imgs := posts[0].Images()
	if len(imgs) != 7 {
		t.Fatalf("фото %d, ожидали 7 (все вложения): %v", len(imgs), imgs)
	}
	// порядок как в посте + из каждой группы взят КРУПНЫЙ размер
	want := []string{
		"https://sun/p1_w.jpg", "https://sun/p2_w.jpg", "https://sun/p3_w.jpg",
		"https://sun/p4_w.jpg", "https://sun/p5_w.jpg", "https://sun/p6_w.jpg",
		"https://sun/p7_w.jpg",
	}
	for i, w := range want {
		if imgs[i] != w {
			t.Errorf("фото %d = %q, ожидали %q", i, imgs[i], w)
		}
	}
	// все фото — ОДНИМ блоком-галереей (клиент рисует сетку, а не 7 одиночных)
	var galleries int
	for _, b := range posts[0].Blocks {
		if b.Kind == BlockImage {
			galleries++
		}
	}
	if galleries != 1 {
		t.Errorf("блоков-галерей %d, ожидали 1", galleries)
	}
}

// ВК-упоминание в посте должно стать ссылкой, а не сырым текстом.
func TestVK_MentionsBecomeLinks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"response":{"count":1,"items":[
		 {"id":1,"owner_id":-204353467,"date":1732000000,
		 "text":"Спасибо [club204353467|СРЕДНЕВЕКОВКА] за праздник!"}]}}`))
	}))
	defer srv.Close()

	old := vkAPIBase
	defer setVKBase(old)
	setVKBase(srv.URL + "/method/")

	posts, err := VK{OwnerID: "204353467", Token: "tok"}.Posts()
	if err != nil {
		t.Fatalf("Posts: %v", err)
	}
	if strings.Contains(posts[0].Text(), "[club") {
		t.Fatalf("сырая разметка в тексте: %q", posts[0].Text())
	}
	links := posts[0].Links()
	if len(links) != 1 || links[0].Href != "https://vk.com/club204353467" {
		t.Fatalf("ссылка не собралась: %+v", links)
	}
}
