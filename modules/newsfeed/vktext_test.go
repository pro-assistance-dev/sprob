package newsfeed

import (
	"strings"
	"testing"
)

// ⚠️ Реальные строки из ленты @srednevekovka (20.11.2026).

// Голый адрес ВК должен стать ссылкой; «прилипшее» слово — отрезаться.
func TestSplitBareURLs_RealAlbumLink(t *testing.T) {
	text := "СМОТРИТЕ В АЛЬБОМЕ! https://vk.ru/album-119823711_311911110 Спасибо огромное Даниилу"
	clean, links := SplitBareURLs(text)

	if len(links) != 1 {
		t.Fatalf("ссылок %d: %+v", len(links), links)
	}
	if links[0].Href != "https://vk.ru/album-119823711_311911110" {
		t.Errorf("адрес: %q", links[0].Href)
	}
	if strings.Contains(clean, "311911110Спасибо") {
		t.Errorf("слово прилипло к адресу: %q", clean)
	}
	if !strings.Contains(clean, "Спасибо огромное") {
		t.Errorf("текст после ссылки потерялся: %q", clean)
	}
}

// ВК склеивает адрес и слово БЕЗ пробела: «…311911113Спасибо».
func TestSplitBareURLs_GluedCyrillic(t *testing.T) {
	clean, links := SplitBareURLs("Смотрите https://vk.ru/album-1_2Спасибо всем")
	if len(links) != 1 || links[0].Href != "https://vk.ru/album-1_2" {
		t.Fatalf("ссылка: %+v", links)
	}
	if !strings.Contains(clean, "Спасибо всем") {
		t.Errorf("слово потерялось: %q", clean)
	}
}

// Хвостовая пунктуация не входит в адрес.
func TestSplitBareURLs_TrailingPunctuation(t *testing.T) {
	_, links := SplitBareURLs("Подробнее: https://vk.ru/wall-1_2.")
	if len(links) != 1 || links[0].Href != "https://vk.ru/wall-1_2" {
		t.Fatalf("точка попала в адрес: %+v", links)
	}
}

// Без ссылок текст не меняется.
func TestSplitBareURLs_None(t *testing.T) {
	clean, links := SplitBareURLs("Обычный текст без адресов")
	if clean != "Обычный текст без адресов" || len(links) != 0 {
		t.Fatalf("%q %+v", clean, links)
	}
}

// ⚠️ Реальный хвост поста: строка из одних тегов — удаляется целиком.
func TestStripHashtags_RealLine(t *testing.T) {
	text := "Приходите на файер-шоу!\n\n#средневековка2026 #фаершоу #фаерщицы #огонь #фаяры #атмосферное"
	got := StripHashtags(text)
	if strings.Contains(got, "#") {
		t.Fatalf("хештеги остались: %q", got)
	}
	if !strings.Contains(got, "Приходите на файер-шоу!") {
		t.Errorf("текст потерялся: %q", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Errorf("пустой абзац в конце: %q", got)
	}
}

// Повторяющиеся теги с подчёркиванием (реальная строка из другого поста).
func TestStripHashtags_UnderscoreTags(t *testing.T) {
	got := StripHashtags("Текст поста.\n\n#КвК #КвК_Ярмарка #КвК_Фестиваль")
	if strings.Contains(got, "#") {
		t.Fatalf("хештеги остались: %q", got)
	}
	if !strings.Contains(got, "Текст поста.") {
		t.Errorf("текст потерялся: %q", got)
	}
}

// Тег ВНУТРИ фразы: слова не должны склеиться.
func TestStripHashtags_InlineTag(t *testing.T) {
	got := StripHashtags("Приезжайте к нам #средневековка2026 весной")
	if strings.Contains(got, "#") {
		t.Fatalf("тег остался: %q", got)
	}
	if !strings.Contains(got, "к нам") || !strings.Contains(got, "весной") {
		t.Errorf("слова склеились: %q", got)
	}
}

// «#1» (номер дома) НЕ является тегом и не вырезается.
func TestStripHashtags_KeepsNumbers(t *testing.T) {
	got := StripHashtags("Ждём вас: дом #1, поле #2")
	if !strings.Contains(got, "#1") || !strings.Contains(got, "#2") {
		t.Errorf("вырезали номер: %q", got)
	}
}
