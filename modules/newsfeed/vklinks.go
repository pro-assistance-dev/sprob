package newsfeed

import (
	"regexp"
	"strings"
)

// Разбор ВК-разметки в тексте поста.
//
// ⚠️ ЗАЧЕМ (20.11.2026, замечание владельца). ВКонтакте пишет упоминания
// сообществ и людей в виде `[club204353467|СРЕДНЕВЕКОВКА]`, `[id519984931|Марьяна
// Ерофеева]`. В API это просто ТЕКСТ — никаких `<a>`. На сайте sred такие
// «ссылки» выводились как есть, сырым текстом со скобками и вертикальной чертой.
// В нашей ленте (@srednevekovka) их десятки: 30 упоминаний своего же клуба,
// 8 — партнёров, 20+ имён людей.
//
// Что здесь: превращение `[owner|подпись]` в Link с настоящим адресом и
// подписью. Сама подстановка в блоки — в `SplitVKMentions`, которую зовут
// источники `VK` и `Telegram` после очистки текста.

// vkMentionRe — `[club<id>|подпись]`, `[id<id>|подпись]`, `[public<id>|…]`,
// `[event<id>|…]`, `[<screen_name>|…]`, `[vk.com/…|…]`.
//
// ⚠️ Подпись может содержать `|` и скобки (имя с чертой, название с эмодзи),
// поэтому «до первой `]`», а не «до первой `|`»: `[^\]]*` и разрезание по
// ПЕРВОЙ черте внутри уже найденной строки.
var vkMentionRe = regexp.MustCompile(`\[([a-zA-Z0-9_.]+)\|([^\]]{0,200})\]`)

// SplitVKMentions — находит ВК-упоминания и отдаёт текст + ссылки.
//
// Возвращает очищенный текст (скобки и служебный id заменены подписью) и
// список ссылок В ПОРЯДКЕ ПОЯВЛЕНИЯ. Так их кладёт в блок `Links`: клиент
// отрисует настоящие `<a>`, а не сырую разметку.
//
// Пример: `Спасибо [club204353467|СРЕДНЕВЕКОВКА]!` →
//
//	текст: `Спасибо СРЕДНЕВЕКОВКА!`
//	ссылки: [{Text: "СРЕДНЕВЕКОВКА", Href: "https://vk.com/club204353467"}]
//
// ⚠️ Упоминание, которое РАВНО тексту поста целиком, тоже разбирается: в
// коротких постах автора упоминают одной строкой, и без этого блок остался бы
// пустым (текст == подпись → вычлось «содержимое поста пусто»).
func SplitVKMentions(text string) (string, []Link) {
	if text == "" || !strings.Contains(text, "[") {
		return text, nil
	}
	var links []Link
	clean := vkMentionRe.ReplaceAllStringFunc(text, func(m string) string {
		sub := vkMentionRe.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		owner, label := sub[1], sub[2]
		href := VKProfileURL(owner)
		if href == "" {
			return m // не наш формат — оставляем как есть, не выдумываем
		}
		label = strings.TrimSpace(label)
		if label == "" {
			label = owner // подписи нет — покажем сам id/имя как в ВК
		}
		links = append(links, Link{Text: label, Href: href})
		return label
	})
	return clean, links
}

// VKProfileURL — адрес страницы по owner из ВК-разметки:
// `club…` → vk.com/club…, `id…` → vk.com/id…, `public…` → vk.com/public…,
// экранное имя → vk.com/<имя>.
//
// ⚠️ `id519984931` — это ЛИЧНАЯ страница, и адрес у неё `vk.com/id519984931`
// (не `vk.com/519984931`): в ВК оба открываются, но канонический — с `id`.
func VKProfileURL(owner string) string {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return ""
	}
	// служебные префиксы оставляем как есть
	low := strings.ToLower(owner)
	for _, p := range []string{"club", "public", "event", "id"} {
		if strings.HasPrefix(low, p) {
			return "https://vk.com/" + owner
		}
	}
	// экранное имя (латиница, цифры, подчёркивание, точка) — это страница/группа
	if ok, _ := regexp.MatchString(`^[A-Za-z][A-Za-z0-9_.]{1,31}$`, owner); ok {
		return "https://vk.com/" + owner
	}
	return ""
}
