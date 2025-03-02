package basehandler

import (
	"github.com/pro-assistance-dev/sprob/helper"
	"github.com/uptrace/bun"
)

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
	t        T
	helper   *helper.Helper
	relation func(*bun.SelectQuery) *bun.SelectQuery
}

func InitR[T Relationable]() Repository[T] {
	r := Repository[T]{helper: Helper}
	return r
}

func rel[T Relationable](x T) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return x.Relation(q)
	}
}

func InitH[T Relationable]() Handler[T] {
	handler := Handler[T]{helper: Helper}
	r := InitR[T]()
	t := Str[T]{}
	r.relation = t.genericValue.Relation
	handler.S = InitS[T](r)
	return handler
}

type Str[T Relationable] struct {
	genericValue T
}

func InitS[T Relationable](r Repository[T]) Service[T] {
	s := Service[T]{helper: Helper}
	s.R = r
	return s
}

func Init[T Relationable]() Handler[T] {
	he := InitH[T]()
	return he
}
