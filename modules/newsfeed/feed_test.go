package newsfeed

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// fileServer отдаёт файл ленты по HTTP — тест гоняет НАСТОЯЩИЙ путь
// (Fetch.Get → fromRSS), а не отдельно взятый парсер.
func fileServer(t *testing.T, path string) *httptest.Server {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("нет файла ленты %s", path)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(b)
	}))
}

// ⚠️ Живая лента @darbore_city (20.11.2026): проверяем то, что реально
// ломалось — посты без ссылок, без картинок, с датой «сейчас».
func TestTelegram_RealFeed(t *testing.T) {
	srv := fileServer(t, "/tmp/darbore-full.xml")
	defer srv.Close()

	tt := Telegram{RSSURL: srv.URL}
	posts, err := tt.fromRSS()
	if err != nil {
		t.Fatalf("fromRSS: %v", err)
	}
	if len(posts) == 0 {
		t.Fatal("лента разобралась в пустоту")
	}

	var withLinks, withImages int
	ids := map[string]bool{}
	for _, p := range posts {
		if ids[p.ID] {
			t.Errorf("дубль поста id=%s", p.ID)
		}
		ids[p.ID] = true
		if len(p.Links()) > 0 {
			withLinks++
		}
		if len(p.Images()) > 0 {
			withImages++
		}
		if p.PublishedAt.IsZero() {
			t.Errorf("пост %s без даты публикации", p.ID)
		}
		if p.URL == "" {
			t.Errorf("пост %s без адреса", p.ID)
		}
	}
	if withLinks == 0 {
		t.Error("ни в одном посте не нашлось ссылок — они теряются при разборе")
	}
	if withImages == 0 {
		t.Error("ни в одном посте не нашлось картинок")
	}
	t.Logf("постов=%d, со ссылками=%d, с картинками=%d", len(posts), withLinks, withImages)

	// ссылки должны быть настоящими адресами, а не подписями без href
	for _, p := range posts[:min(3, len(posts))] {
		for _, l := range p.Links() {
			if l.Text == "" || l.Href == "" {
				t.Errorf("пост %s: ссылка без подписи или адреса: %+v", p.ID, l)
			}
		}
	}
}

// Дата берётся ИЗ ПОСТА, а не «сейчас»: иначе хронология врёт при импорте.
func TestParsePubDate(t *testing.T) {
	cases := map[string]string{
		"Fri, 20 Nov 2026 16:09:32 GMT":   "2026-11-20T16:09:32Z",
		"Fri, 20 Nov 2026 16:09:32 +0300": "2026-11-20T13:09:32Z",
		"2026-11-20T16:09:32Z":            "2026-11-20T16:09:32Z",
	}
	for in, want := range cases {
		got := parsePubDate(in)
		if got.IsZero() {
			t.Errorf("%q: дата не разобралась", in)
			continue
		}
		if got.Format("2006-01-02T15:04:05Z") != want {
			t.Errorf("%q → %s, ожидали %s", in, got.Format("2006-01-02T15:04:05Z"), want)
		}
	}
	if !parsePubDate("не дата").IsZero() {
		t.Error("мусор должен давать нулевую дату")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
