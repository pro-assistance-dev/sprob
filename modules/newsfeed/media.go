package newsfeed

import (
	"regexp"
	"strings"
)

// Разбор медиа и служебных вставок.
//
// ⚠️ ЗАЧЕМ (20.11.2026, замер живой ленты @darbore_city):
// - `vk.com/video`, YouTube, Rutube — встречаются в постах, но НЕ
// встраивались: пост выглядел оборванным (ссылка без плеера);
// - `telesco.pe` — обрезанное RSSHub видео: вставляем НЕЛЬЗЯ, надо честно
// сказать «видео в источнике», а не молча терять;
// - RSSHub вместо видео вставляет `<blockquote><b>Video is too big</b>…`
// — после снятия тегов оставалось «is too big» посреди фразы.

var (
	// Видео-ссылки платформ, которые умеют встраиваться.
	ytRe      = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([A-Za-z0-9_-]{6,})`)
	vkVideoRe = regexp.MustCompile(`vk(?:video\.ru|\.com)/video(-?\d+)_(\d+)`)
	rutubeRe  = regexp.MustCompile(`rutube\.ru/video/([a-f0-9]{32})`)
	// telesco.pe — CDN Telegram: видео оттуда не встроить.
	telescoRe = regexp.MustCompile(`(?i)telesco\.pe|cdn\d*\.telesco\.pe`)

	// Заглушка RSSHub вместо видео (HTML ещё не снят).
	rssVideoBlockquoteRe = regexp.MustCompile(`(?is)<blockquote>\s*<b>\s*Video is too big\s*</b>.*?</blockquote>`)
	orphanVideoRe        = regexp.MustCompile(`(?im)^\s*(?:video\s+)?is\s+too\s+(?:big|large)\.?\s*$`)
)

// CleanArtifacts — убирает служебные вставки RSSHub из HTML ДО снятия тегов.
//
// ⚠️ Порядок важен: если снять теги первыми, от цитаты останется безликое
// «is too big» посреди нормальной фразы (проверено в darbore на постах
// 129/133/134/135; в самом канале этих слов НЕТ).
func CleanArtifacts(raw string) string {
	s := rssVideoBlockquoteRe.ReplaceAllString(raw, " ")
	s = orphanVideoRe.ReplaceAllString(s, "")
	return s
}

// VideoEmbed — разбирает ссылку на видео.
//
// Возвращает: embed (адрес для iframe; пусто — встроить нельзя) и
// провайдера-подпись для клиента.
func VideoEmbed(href string) (embed, provider string) {
	if m := ytRe.FindStringSubmatch(href); len(m) > 1 {
		return "https://www.youtube.com/embed/" + m[1], "youtube"
	}
	if m := vkVideoRe.FindStringSubmatch(href); len(m) > 2 {
		// VK: https://vk.com/video_ext.php?oid=<owner>&id=<video>
		return "https://vk.com/video_ext.php?oid=" + m[1] + "&id=" + m[2], "vk"
	}
	if m := rutubeRe.FindStringSubmatch(href); len(m) > 1 {
		return "https://rutube.ru/play/embed/" + m[1], "rutube"
	}
	if telescoRe.MatchString(href) {
		return "", "telegram"
	}
	return "", ""
}

// IsVideoLink — стоит ли считать ссылку видео (клиент рисует плеер/заглушку).
func IsVideoLink(href string) bool {
	embed, provider := VideoEmbed(href)
	return embed != "" || provider != ""
}

// GroupGallery — собирает подряд идущие одиночные картинки в один блок
// (медиагруппа Telegram приходит несколькими <img> подряд).
//
// Зачем: клиенту удобнее одна карусель, чем N отдельных картинок; и в БД
// проектов это обычно одно поле-массив.
func GroupGallery(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Kind == BlockImage && len(out) > 0 {
			last := &out[len(out)-1]
			if last.Kind == BlockImage {
				last.Images = append(last.Images, b.Images...)
				continue
			}
		}
		out = append(out, b)
	}
	return out
}

// SkipText — служебные посты канала, которым не место в ленте
// («Фото канала обновлено» и подобное).
func SkipText(text string) bool {
	t := strings.ToLower(text)
	for _, s := range []string{
		"channel photo updated",
		"channel photo removed",
		"channel photo changed",
		"channel created",
		"фото канала обновлено",
		"канал создан",
	} {
		if strings.Contains(t, s) {
			return true
		}
	}
	return false
}
