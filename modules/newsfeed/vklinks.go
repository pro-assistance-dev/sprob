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
var vkMentionRe = regexp.MustCompile(`\[([^\]|]{1,200})\|([^\]]{0,200})\]`)

// vkURLMentionRe — `[https://vk.ru/@slug|подпись]`: в ВК так оформляют ссылки
// на статьи и внешние страницы. Отдельная регулярка, потому что owner здесь —
// полный URL с `:` и `/`, а не имя/id.
var vkURLMentionRe = regexp.MustCompile(`\[(https?://[^\]|]{4,300})\|([^\]]{0,200})\]`)

// vkBadgeRe — `[Ссылка]` — подпись-плашка ВК без адреса. Сама по себе она
// бесполезна (некуда вести), но в тексте выглядит мусором: «[Ссылка] Студия…».
// Убираем, оставляя текст вокруг.
var vkBadgeRe = regexp.MustCompile(`\[(?:Ссылка|Фото|Видео|Опрос|Файл|Ссылка\s+удалена)\]\s*`)

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

	// 1) ссылки-статьи `[https://…|подпись]` — до общей регулярки, иначе owner
	// с `:` и `/` в неё не попадёт.
	clean := vkURLMentionRe.ReplaceAllStringFunc(text, func(m string) string {
		sub := vkURLMentionRe.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		href, label := strings.TrimSpace(sub[1]), strings.TrimSpace(sub[2])
		if href == "" {
			return m
		}
		if label == "" {
			label = href
		}
		links = append(links, Link{Text: label, Href: href})
		return label
	})

	// 2) плашки без адреса — просто мусор в тексте.
	clean = vkBadgeRe.ReplaceAllString(clean, "")

	// 3) упоминания `[club123|Подпись]` / `[id123|Имя]`.
	clean = vkMentionRe.ReplaceAllStringFunc(clean, func(m string) string {
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
