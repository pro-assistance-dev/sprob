package newsfeed

import (
	"regexp"
	"strings"
)

// Текст ВК: голые ссылки и хештеги.
//
// ⚠️ ЗАЧЕМ (20.11.2026, замечания владельца):
// 1. «https://vk.ru/album-119823711_311911110 — ссылки должны быть кликабельными».
// ВКонтакте отдаёт адреса ГОЛЫМ ТЕКСТОМ (не `[url|подпись]`), и на сайте они
// не нажимались: человек видел длинную строку и не мог открыть альбом.
// 2. «#средневековка2026 #фаершоу … — это теги, они должны быть удалены».
// Хештеги в ВК — не часть текста, а метки. В ленте их 100 постов подряд,
// и они висели хвостом в каждой новости.
//
// Обе функции чистые (текст → текст + ссылки) и применяются в `VK.convert`
// и `vkimport.normalizeBody` — то есть и для новых постов, и при перепарсинге.

var (
	// bareURLRe — адрес без разметки. Ловим только http(s), чтобы не съесть
	// «vk.com» в обычной фразе; хвостовую пунктуацию отсекаем отдельно.
	//
	// ⚠️ Класс исключает `"<>` и пробелы: ВК иногда склеивает адрес с текстом
	// без пробела («…311911113Спасибо»), поэтому в конце стоит проверка на
	// «мусорный хвост» из кириллицы — см. trimURLTail.
	bareURLRe = regexp.MustCompile(`https?://[^\s<>"']+`)

	// vkHashtagRe — хештег: `#слово` с буквами/цифрами/подчёркиванием.
	// ⚠️ Требуем хотя бы одну букву (`[\p{L}_]` в начале после #): иначе «#1»
	// (номер дома/место) считался бы тегом и вырезался из текста.
	vkHashtagRe = regexp.MustCompile(`(?:^|\s)#[\p{L}_][\p{L}\p{N}_-]*`)

	// hashtagOnlyLineRe — строка, состоящая ТОЛЬКО из хештегов и пробелов
	// («#КвК #КвК_Ярмарка #КвК_Фестиваль»). Такие строки убираем целиком,
	// иначе на их месте остаётся пустой абзац с отступом.
	hashtagOnlyLineRe = regexp.MustCompile(`^\s*(?:#[\p{L}_][\p{L}\p{N}_-]*\s*)+$`)
)

// SplitBareURLs — находит голые ссылки в тексте.
//
// Возвращает текст (ссылки остаются в нём как есть — они читаемы) и список
// `Link` для рендера настоящими `<a>`. Порядок ссылок = порядок в тексте.
func SplitBareURLs(text string) (string, []Link) {
	if text == "" || !strings.Contains(text, "http") {
		return text, nil
	}
	var links []Link
	seen := make(map[string]bool, 4)
	clean := bareURLRe.ReplaceAllStringFunc(text, func(raw string) string {
		u := trimURLTail(raw)
		if u == "" {
			return raw
		}
		// подпись — сам адрес, но без протокола: короче и не ломает строку
		label := strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
		if !seen[u] {
			seen[u] = true
			links = append(links, Link{Text: label, Href: u})
		}
		// ⚠️ В ТЕКСТЕ оставляем САМ АДРЕС, а «прилипший» хвост (слово без
		// пробела) — отдаём обратно в текст: `raw` = u + tail, и именно tail
		// содержал начало слова («Спасибо»). Раньше здесь стоял
		// `strings.Replace(raw, TrimPrefix(raw, u), "")` — он вырезал ХВОСТ,
		// то есть терял слово (поймано тестом 20.11).
		if strings.HasPrefix(raw, u) {
			// возвращаем адрес + хвост (хвост — начало прилипшего слова)
			return raw
		}
		return u
	})
	return clean, links
}

// trimURLTail — убирает прилипшую кириллицу и хвостовую пунктуацию.
//
// ⚠️ Реальный случай ВК: «…/album-119823711_311911113Спасибо огромное».
// Адрес склеен со словом без пробела. Кириллица в URL кодируется и в таком
// виде в адрес не входит никогда — обрезаем по первому её символу.
func trimURLTail(raw string) string {
	u := raw
	for i, r := range u {
		if r >= 'А' && r <= 'я' || r == 'Ё' || r == 'ё' {
			u = u[:i]
			break
		}
	}
	// хвостовая пунктуация
	u = strings.TrimRight(u, ".,;:!?)]»\"'")
	return u
}

// StripHashtags — убирает хештеги из текста.
//
// Строка из одних тегов удаляется целиком (не оставляя пустой абзац);
// тег внутри фразы заменяется пробелом, чтобы слова не склеились.
func StripHashtags(text string) string {
	if text == "" || !strings.Contains(text, "#") {
		return text
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		if hashtagOnlyLineRe.MatchString(ln) {
			continue // строка-теги: убираем целиком
		}
		// одиночный тег в конце/середине строки
		ln = vkHashtagRe.ReplaceAllStringFunc(ln, func(m string) string {
			// сохраняем ведущий пробел, чтобы слова не склеились
			if strings.HasPrefix(m, " ") || strings.HasPrefix(m, "\t") {
				return " "
			}
			return ""
		})
		ln = strings.Join(strings.Fields(ln), " ")
		out = append(out, ln)
	}
	// схлопываем пустые строки, появившиеся от удалённых строк-тегов
	res := strings.Join(out, "\n")
	for strings.Contains(res, "\n\n\n") {
		res = strings.ReplaceAll(res, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(res)
}
