package metabase

import (
	"github.com/pro-assistance-dev/sprob/helper"
)

type Handler struct {
	helper *helper.Helper
	cards  Cards
}

var H *Handler

type Card struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Cards []*Card

func (c Cards) Find(name string) *Card {
	for _, card := range c {
		if card.Name == name {
			return card
		}
	}
	return nil
}

// Init — возвращает handler (Т7). Кэш карточек — поле экземпляра, а не пакетный
// глобал; глобал H заполняется для совместимости.
func Init(h *helper.Helper) *Handler {
	handler := &Handler{helper: h, cards: make(Cards, 0)}
	H = handler
	return handler
}
