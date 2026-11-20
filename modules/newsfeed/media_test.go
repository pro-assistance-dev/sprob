package newsfeed

import "testing"

func TestVideoEmbed(t *testing.T) {
	cases := []struct{ in, embed, provider string }{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "https://www.youtube.com/embed/dQw4w9WgXcQ", "youtube"},
		{"https://youtu.be/dQw4w9WgXcQ", "https://www.youtube.com/embed/dQw4w9WgXcQ", "youtube"},
		{"https://vk.com/video-12345_67890", "https://vk.com/video_ext.php?oid=-12345&id=67890", "vk"},
		{"https://vkvideo.ru/video-12345_67890", "https://vk.com/video_ext.php?oid=-12345&id=67890", "vk"},
		{"https://cdn4.telesco.pe/file/abc.mp4", "", "telegram"},
		{"https://ozon.ru/product/1", "", ""},
	}
	for _, c := range cases {
		embed, provider := VideoEmbed(c.in)
		if embed != c.embed || provider != c.provider {
			t.Errorf("%s → (%q,%q), ожидали (%q,%q)", c.in, embed, provider, c.embed, c.provider)
		}
	}
}

// Медиагруппа: подряд идущие картинки собираются в один блок (карусель).
func TestGroupGallery(t *testing.T) {
	in := []Block{
		{Kind: BlockText, Text: "подпись"},
		{Kind: BlockImage, Images: []string{"a.jpg"}},
		{Kind: BlockImage, Images: []string{"b.jpg"}},
		{Kind: BlockImage, Images: []string{"c.jpg"}},
		{Kind: BlockText, Text: "ещё"},
	}
	out := GroupGallery(in)
	if len(out) != 3 {
		t.Fatalf("ожидали 3 блока, получили %d: %+v", len(out), out)
	}
	if len(out[1].Images) != 3 {
		t.Fatalf("галерея должна собрать 3 картинки, получили %v", out[1].Images)
	}
}

// «Video is too big» — служебная вставка RSSHub, не часть текста поста.
func TestCleanArtifacts_VideoStub(t *testing.T) {
	raw := `Начало.<blockquote><b>Video is too big</b><br><img src="x.jpg"></blockquote>Конец.`
	got := ParseHTML(CleanArtifacts(raw))
	text := Post{Blocks: got}.Text()
	if contains(text, "too big") {
		t.Fatalf("заглушка осталась в тексте: %q", text)
	}
	if !contains(text, "Начало.") || !contains(text, "Конец.") {
		t.Fatalf("текст потерян: %q", text)
	}
}

func TestSkipText_ServicePosts(t *testing.T) {
	for _, s := range []string{"Channel photo updated", "Фото канала обновлено"} {
		if !SkipText(s) {
			t.Errorf("%q должен отбрасываться", s)
		}
	}
	if SkipText("Вышел новый выпуск комикса") {
		t.Error("обычный пост отбрасываться не должен")
	}
}
