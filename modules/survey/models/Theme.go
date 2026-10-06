package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Theme — тема оформления формы, показываемая респонденту при заполнении.
// Хранит набор дизайн-токенов (CSS custom properties: цвета, радиус, шрифт),
// которые рантайм навешивает на корень опроса. Это НЕ произвольный CSS
// (безопасность): только декларативные значения переменных.
type Theme struct {
	bun.BaseModel `bun:"themes,alias:themes"`
	ID            uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`

	Name string `json:"name"`
	Slug string `json:"slug"`
	// Tokens — карта CSS-переменных → значений, например
	// {"--brand-surface":"#ffffff","--button-primary-background":"#476db5","--border-radius-default":"12px"}.
	Tokens map[string]string `bun:"tokens,type:jsonb" json:"tokens"`
	// Dark — тема рассчитана на тёмный фон (для контрастной подписи и превью).
	Dark bool `json:"dark"`

	SortOrder int `bun:"sort_order" json:"sortOrder"`

	CreatedAt time.Time `bun:"created_at" json:"createdAt"`
}

type Themes []*Theme

type ThemesWithCount struct {
	Themes Themes `json:"items"`
	Count  int    `json:"count"`
}

func (item Theme) Relation(q *bun.SelectQuery) *bun.SelectQuery {
	return q
}

// AllowedTokenPrefix — белый список префиксов CSS-переменных темы.
// Рантайм применяет только ключи из белого списка (защита от инъекций).
func IsAllowedToken(key string) bool {
	return len(key) > 2 && key[0] == '-' && key[1] == '-'
}
