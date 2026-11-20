// Package newsfeed — транспорт ленты новостей: источник → структурированный пост.
//
// ЗАЧЕМ ОТДЕЛЬНЫМ ПАКЕТОМ (20.11.2026). Парсинг ленты был написан ДВАЖДЫ —
// в ferma и darbore (5–6 файлов, 574 и 980 строк), при том что источники и
// библиотеки одинаковые. sred берёт ленту из ВКонтакте — третий почти такой же
// код. Копии разошлись: фикс обрыва RSSHub жил только в darbore.
//
// ГРАНИЦА ОТВЕТСТВЕННОСТИ (важно, чтобы не превратиться в «модуль всё»):
//   - ЗДЕСЬ: сходить в источник, разобрать пост, скачать медиа, почистить текст,
//     отдать структуру для показа;
//   - В ПРОЕКТЕ: модель и схема БД, автопубликация, комментарии, заглушки,
//     редиректы — у ferma и darbore они разные, и сводить их в общий пакет
//     нельзя (см. TECH_DEBT О33).
//
// Пост отдаётся НЕ плоской строкой, а блоками (см. Block): иначе ссылки и
// картинки теряются при «просто сними теги» — проверено 20.11 на живом канале:
// «Слушать Яндекс Музыка» оставалось текстом, на который нельзя нажать.
package newsfeed

import "time"

// BlockKind — тип фрагмента поста. Клиент рендерит по нему.
type BlockKind string

const (
	// BlockText — обычный абзац. Inline внутри может содержать ссылки.
	BlockText BlockKind = "text"
	// BlockLink — самостоятельная ссылка-кнопка (например, «Яндекс Музыка»).
	// Нужна, когда ссылка — отдельный элемент подачи, а не слово в фразе.
	BlockLink BlockKind = "link"
	// BlockImage — картинка поста (галерея = несколько блоков подряд).
	BlockImage BlockKind = "image"
	// BlockVideo — встраиваемое видео (провайдер + id) либо ссылка на источник,
	// если встроить нельзя (см. Block.Embed == "").
	BlockVideo BlockKind = "video"
	// BlockQuote — цитата/врезка (<blockquote> в исходнике).
	BlockQuote BlockKind = "quote"
	// BlockList — маркированный список (пункты в Text через \n).
	BlockList BlockKind = "list"
)

// Link — ссылка внутри текста: то, что показывается, и куда ведёт.
// Хранится ОТДЕЛЬНО от текста, чтобы клиент сделал настоящий <a>, а не
// выводил безжизненную подпись (главный баг 20.11).
type Link struct {
	Text string `json:"text"` // видимая подпись: «Яндекс Музыка»
	Href string `json:"href"` // адрес назначения
}

// Block — фрагмент поста.
type Block struct {
	Kind BlockKind `json:"kind"`
	// Text — содержимое для text/quote/list.
	Text string `json:"text,omitempty"`
	// Links — ссылки, встреченные ВНУТРИ текста этого блока, по порядку.
	Links []Link `json:"links,omitempty"`
	// Images — адреса картинок блока (после публикации — локальные).
	Images []string `json:"images,omitempty"`
	// Embed — для video: готовый адрес iframe (может быть пустым, если
	// провайдер не поддерживает встраивание — тогда VideoURL даёт ссылку).
	Embed string `json:"embed,omitempty"`
	// VideoURL — исходная ссылка на видео (всегда заполняется для video).
	VideoURL string `json:"videoUrl,omitempty"`
}

// Post — один пост источника, приведённый к общему виду.
//
// Все поля уже «чистые»: текст без служебной оболочки канала, картинки —
// абсолютные адреса (скачивание делает проект: ему решать, где хранить файлы).
type Post struct {
	// ID — идентификатор поста В ИСТОЧНИКЕ (номер в Telegram, id записи в ВК).
	// Уникальный ключ дедупликации: проекты пишут его в свою таблицу.
	ID string `json:"id"`
	// Title — заголовок поста, если источник его отдаёт. Пусто у постов,
	// где заголовок не предусмотрен (Telegram чаще всего).
	Title string `json:"title,omitempty"`
	// Blocks — содержимое поста по порядку.
	Blocks []Block `json:"blocks"`
	// PublishedAt — когда пост ОПУБЛИКОВАН в источнике (не когда мы его
	// увидели: иначе вся лента при первом импорте получает одну дату и
	// хронология врёт — баг ferma, замечен 20.11).
	PublishedAt time.Time `json:"publishedAt"`
	// URL — постоянный адрес поста в источнике (для «открыть оригинал»).
	URL string `json:"url,omitempty"`
	// Edited — пост правили после публикации (источник это сообщает).
	Edited bool `json:"edited,omitempty"`
}

// Text — плоский текст поста: все текстовые блоки через пустую строку.
// Нужен проектам, которые пока хранят текст одной строкой (ферма).
func (p Post) Text() string {
	out := ""
	for _, b := range p.Blocks {
		switch b.Kind {
		case BlockText, BlockQuote, BlockList:
			if b.Text == "" {
				continue
			}
			if out != "" {
				out += "\n\n"
			}
			out += b.Text
		}
	}
	return out
}

// Images — все картинки поста по порядку.
func (p Post) Images() []string {
	var out []string
	for _, b := range p.Blocks {
		out = append(out, b.Images...)
	}
	return out
}

// Links — все ссылки поста по порядку (и самостоятельные, и внутри текста).
func (p Post) Links() []Link {
	var out []Link
	for _, b := range p.Blocks {
		if b.Kind == BlockLink && b.Embed == "" && b.Text != "" && len(b.Links) == 1 {
			out = append(out, b.Links...)
			continue
		}
		out = append(out, b.Links...)
	}
	return out
}
