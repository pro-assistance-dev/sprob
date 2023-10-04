package basehandler

import (
	"context"

	"github.com/google/uuid"
)

type IServiceWithContext[TSingle, TPlural, TPluralWithCount any] interface {
	Create(context.Context, *TSingle) error
	GetAll(context.Context) (TPluralWithCount, error)
	Get(context.Context, string) (*TSingle, error)
	Delete(context.Context, string) error
	Update(context.Context, *TSingle) error
}

type IServiceWithManyWithContext[TSingle, TPlural, TPluralWithCount any] interface {
	IServiceWithContext[TSingle, TPlural, TPluralWithCount]
	UpsertMany(context.Context, TPlural) error
	DeleteMany(context.Context, []uuid.UUID) error
}

type IRepositoryWithContext[TSingle, TPlural, TPluralWithCount any] interface {
	Create(context.Context, *TSingle) error
	Update(context.Context, *TSingle) error
	GetAll(context.Context) (TPluralWithCount, error)
	Get(context.Context, string) (*TSingle, error)
	Delete(context.Context, string) error
}

type IRepositoryWithManyWithContext[TSingle, TPlural, TPluralWithCount any] interface {
	IRepositoryWithContext[TSingle, TPlural, TPluralWithCount]
	Upsert(context.Context, *TSingle) error
	UpsertMany(context.Context, TPlural) error
	DeleteMany(context.Context, []uuid.UUID) error
}
