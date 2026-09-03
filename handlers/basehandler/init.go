package basehandler

import (
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/uptrace/bun"
)

// Helper — legacy-глобал (service locator): заполняется SetHelper при старте
// приложения (routing/router.go потребителей) и используется Init*/InitR по
// умолчанию. Новый код — конструкторы NewR/NewS/NewH с явным helper
// (тесты и routing.WithHelper, Т7): глобал не требуется.
var Helper *helper.Helper

func SetHelper(h *helper.Helper) {
	Helper = h
}

type Relationable interface {
	Relation(*bun.SelectQuery) *bun.SelectQuery
}

type Handler[T Relationable] struct {
	S      Service[T]
	helper *helper.Helper
}

type Service[T Relationable] struct {
	R      Repository[T]
	helper *helper.Helper
}

type Repository[T Relationable] struct {
	helper   *helper.Helper
	relation func(*bun.SelectQuery) *bun.SelectQuery
}

// NewR — repository с явным helper. Relation не задаётся (полный handler — NewH).
func NewR[T Relationable](h *helper.Helper) Repository[T] {
	return Repository[T]{helper: h}
}

// NewS — service с явным helper и repository.
func NewS[T Relationable](h *helper.Helper, r Repository[T]) Service[T] {
	return Service[T]{helper: h, R: r}
}

// NewH — полный handler (repository + relation + service) с явным helper.
func NewH[T Relationable](h *helper.Helper) Handler[T] {
	r := NewR[T](h)
	t := Str[T]{}
	r.relation = t.genericValue.Relation
	return Handler[T]{helper: h, S: NewS[T](h, r)}
}

// InitR/InitS/InitH/Init — legacy-обёртки над New*: используют глобал Helper
// (совместимость; новые вызовы — через New*/WithHelper).
func InitR[T Relationable]() Repository[T] {
	return NewR[T](Helper)
}

func InitS[T Relationable](r Repository[T]) Service[T] {
	return NewS[T](Helper, r)
}

func InitH[T Relationable]() Handler[T] {
	return NewH[T](Helper)
}

func Init[T Relationable]() Handler[T] {
	return InitH[T]()
}

type Str[T Relationable] struct {
	genericValue T
}
