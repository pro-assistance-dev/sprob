package metabase

import (
	"github.com/pro-assistance-dev/sprob/helper"
)

type Handler struct {
	helper *helper.Helper
}

var H *Handler

type Card struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Cards []*Card

var cards = make(Cards, 0)

func (c Cards) Find(name string) *Card {
	for _, card := range c {
		if card.Name == name {
			return card
		}
	}
	return nil
}

func Init(h *helper.Helper) {
	H = &Handler{helper: h}
	// H.Cards()
}
