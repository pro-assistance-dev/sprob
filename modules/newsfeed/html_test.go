package newsfeed

import (
	"testing"
)

// ⚠️ Главный тест пакета (20.11.2026). Основан на РЕАЛЬНОМ куске поста из
// канала @darbore_city: раньше из него получался мёртвый текст
// «Слушать Яндекс Музыка | VK Музыка» — подписи без адресов.
const realPost = `<p>Полная аудиоверсия первой книги &quot;Исповеди в жанре нуар&quot;: 53 главы, около 5 часов.<br><br><b>Слушать</b><br><a href="https://music.yandex.ru/album/38166701" target="_blank">Яндекс Музыка</a> | <a href="https://vk.ru/artist/9186651729207388728" target="_blank">VK Музыка</a><br><br>Аллен Хёрт - бывший детектив наркоконтроля.</p><img src="https://cdn4.telesco.pe/file/abc.jpg" width="800">`

func TestParseHTML_KeepsLinksFromRealPost(t *testing.T) {
	blocks := ParseHTML(realPost)
	if len(blocks) == 0 {
		t.Fatal("пост разобрался в пустоту")
	}

	links := Post{Blocks: blocks}.Links()
	if len(links) < 2 {
		t.Fatalf("ожидали 2+ ссылки, получили %d: %+v", len(links), links)
	}

	want := map[string]string{
		"Яндекс Музыка": "https://music.yandex.ru/album/38166701",
		"VK Музыка": "https://vk.ru/artist/9186651729207388728",
	}
	for _, l := range links {
		if href, ok := want[l.Text]; ok && l.Href != href {
			t.Errorf("ссылка %q ведёт на %q, ожидали %q", l.Text, l.Href, href)
		}
	}

	// картинка должна быть отдельным блоком и в https
	imgs := Post{Blocks: blocks}.Images()
	if len(imgs) != 1 || imgs[0] != "https://cdn4.telesco.pe/file/abc.jpg" {
		t.Fatalf("картинки: %v", imgs)
	}
}

func TestParseHTML_TextKeepsMeaning(t *testing.T) {
	text := Post{Blocks: ParseHTML(realPost)}.Text()
	for _, want := range []string{"Полная аудиоверсия", "Аллен Хёрт", "Яндекс Музыка"} {
		if !contains(text, want) {
			t.Errorf("в тексте нет %q:\n%s", want, text)
		}
	}
	// и никакого HTML
	if contains(text, "<") {
		t.Errorf("в тексте остались теги:\n%s", text)
	}
}

// Ссылка внутри фразы — Link, а не отдельный блок: иначе фраза развалится.
func TestParseHTML_InlineLink(t *testing.T) {
	blocks := ParseHTML(`<p>Книга есть <a href="https://ozon.ru/x">на Ozon</a> и в магазинах.</p>`)
	if len(blocks) != 1 || blocks[0].Kind != BlockText {
		t.Fatalf("ожидали один текстовый блок, получили %+v", blocks)
	}
	if len(blocks[0].Links) != 1 || blocks[0].Links[0].Href != "https://ozon.ru/x" {
		t.Fatalf("ссылка потерялась: %+v", blocks[0].Links)
	}
	if !contains(blocks[0].Text, "на Ozon") {
		t.Fatalf("подпись ссылки не в тексте: %q", blocks[0].Text)
	}
}

// «Островная» ссылка (Telegram подаёт кнопками построчно) — отдельный блок.
func TestParseHTML_StandaloneLink(t *testing.T) {
	blocks := ParseHTML(`<a href="https://music.yandex.ru/album/1">Яндекс Музыка</a><br><a href="https://vk.ru/artist/2">VK Музыка</a>`)
	var links int
	for _, b := range blocks {
		if b.Kind == BlockLink {
			links++
		}
	}
	if links != 2 {
		t.Fatalf("ожидали 2 кнопки-ссылки, получили %d: %+v", links, blocks)
	}
}

func TestParseHTML_ProtoEqualRelativeURLs(t *testing.T) {
	blocks := ParseHTML(`<img src="//cdn4.telesco.pe/x.jpg">`)
	imgs := Post{Blocks: blocks}.Images()
	if len(imgs) != 1 || imgs[0] != "https://cdn4.telesco.pe/x.jpg" {
		t.Fatalf("// → https не применился: %v", imgs)
	}
}

func TestParseHTML_Empty(t *testing.T) {
	if got := ParseHTML(" \n "); got != nil {
		t.Fatalf("пустой пост должен дать nil, получили %+v", got)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
