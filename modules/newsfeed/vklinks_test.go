package newsfeed

import (
	"strings"
	"testing"
)

// ⚠️ Реальный текст из ленты @srednevekovka (20.11.2026). ВКонтакте пишет
// упоминания как `[club<id>|Подпись]` — это НЕ <a>, а обычный текст в API.
// На сайте они выводились сырыми: «[club204353467|СРЕДНЕВЕКОВКА]».
const mentionText = "Спасибо [club204353467|СРЕДНЕВЕКОВКА —Фестиваль Фэнтези и Реконструкции] " +
	"и лично [id519984931|Марьяна Ерофеева] за помощь!"

func TestSplitVKMentions_RealText(t *testing.T) {
	clean, links := SplitVKMentions(mentionText)

	if strings.Contains(clean, "[") || strings.Contains(clean, "|") {
		t.Fatalf("разметка осталась в тексте: %q", clean)
	}
	if !strings.Contains(clean, "СРЕДНЕВЕКОВКА") || !strings.Contains(clean, "Марьяна Ерофеева") {
		t.Errorf("подписи потерялись: %q", clean)
	}

	if len(links) != 2 {
		t.Fatalf("ссылок %d, ожидали 2: %+v", len(links), links)
	}
	if links[0].Href != "https://vk.com/club204353467" {
		t.Errorf("сообщество: %q", links[0].Href)
	}
	// ⚠️ Личная страница — канонический адрес с префиксом `id`.
	if links[1].Href != "https://vk.com/id519984931" {
		t.Errorf("личная страница: %q", links[1].Href)
	}
	if links[0].Text == "" || links[1].Text == "" {
		t.Errorf("пустые подписи: %+v", links)
	}
}

// Подпись с вертикальной чертой и эмодзи — тоже корректно (до `]`, не до `|`).
func TestSplitVKMentions_PipeInLabel(t *testing.T) {
	clean, links := SplitVKMentions("Смотрите [club201889078|ЯрУльф | Огненное шоу] тут")
	if len(links) != 1 {
		t.Fatalf("ссылок %d: %+v", len(links), links)
	}
	if links[0].Text != "ЯрУльф | Огненное шоу" {
		t.Errorf("подпись обрезалась: %q", links[0].Text)
	}
	if !strings.Contains(clean, "ЯрУльф | Огненное шоу") {
		t.Errorf("текст: %q", clean)
	}
}

// Текст без упоминаний не меняется, ссылок нет.
func TestSplitVKMentions_NoMentions(t *testing.T) {
	clean, links := SplitVKMentions("Обычный пост без ссылок")
	if clean != "Обычный пост без ссылок" || len(links) != 0 {
		t.Fatalf("изменился обычный текст: %q %+v", clean, links)
	}
}

// Неизвестный формат (не ВК-owner) — оставляем как есть, не выдумываем адрес.
func TestSplitVKMentions_UnknownOwner(t *testing.T) {
	clean, links := SplitVKMentions("текст [чтото-странное|Подпись] тут")
	if len(links) != 0 {
		t.Fatalf("выдумали ссылку: %+v", links)
	}
	if !strings.Contains(clean, "[чтото-странное|Подпись]") {
		t.Errorf("неизвестное сохранили неверно: %q", clean)
	}
}

func TestVKProfileURL(t *testing.T) {
	cases := map[string]string{
		"club204353467": "https://vk.com/club204353467",
		"public123":     "https://vk.com/public123",
		"event55":       "https://vk.com/event55",
		"id519984931":   "https://vk.com/id519984931",
		"artfaina":      "https://vk.com/artfaina",
		"":              "",
	}
	for in, want := range cases {
		if got := VKProfileURL(in); got != want {
			t.Errorf("%q → %q, ожидали %q", in, got, want)
		}
	}
}

// ⚠️ Реальные форматы из ленты (20.11.2026): ссылка на статью ВК и плашка
// «[Ссылка]». Первая должна стать ссылкой, вторая — исчезнуть.
func TestSplitVKMentions_ArticleURL(t *testing.T) {
	clean, links := SplitVKMentions("Читайте в [https://vk.ru/@srednevekovka-kak-dobratsya|статье], подробнее там")
	if len(links) != 1 {
		t.Fatalf("ссылок %d: %+v", len(links), links)
	}
	if links[0].Href != "https://vk.ru/@srednevekovka-kak-dobratsya" {
		t.Errorf("адрес: %q", links[0].Href)
	}
	if links[0].Text != "статье" {
		t.Errorf("подпись: %q", links[0].Text)
	}
	if strings.Contains(clean, "[") {
		t.Errorf("разметка осталась: %q", clean)
	}
}

func TestSplitVKMentions_DropsBadge(t *testing.T) {
	clean, _ := SplitVKMentions("[Ссылка] Студия цыганского творчества «Рада»")
	if strings.Contains(clean, "[Ссылка]") {
		t.Errorf("плашка осталась: %q", clean)
	}
	if !strings.Contains(clean, "Студия цыганского") {
		t.Errorf("текст потерялся: %q", clean)
	}
}
