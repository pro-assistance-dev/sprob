package basehandler

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IHandler interface {
	GetAll(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Delete(c *gin.Context)
	Update(c *gin.Context)
}

type IService[TSingle, TPlural, TPluralWithCount any] interface {
	Create(context.Context, *TSingle) error
	GetAll(context.Context) (TPluralWithCount, error)
	Get(context.Context, string) (*TSingle, error)
	Delete(context.Context, string) error
	Update(context.Context, *TSingle) error
}

type IServiceWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IService[TSingle, TPlural, TPluralWithCount]
	UpsertMany(*TPlural) error
	DeleteMany([]uuid.UUID) error
}

type IRepository[TSingle, TPlural, TPluralWithCount any] interface {
	Create(context.Context, *TSingle) error
	Update(context.Context, *TSingle) error
	GetAll(context.Context) (TPluralWithCount, error)
	Get(context.Context, string) (*TSingle, error)
	Delete(context.Context, string) error
}

type IRepositoryWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IRepository[TSingle, TPlural, TPluralWithCount]
	Upsert(context.Context, *TSingle) error
	UpsertMany(context.Context, *TPlural) error
	DeleteMany(context.Context, []uuid.UUID) error
}

type IFilesService interface {
	Upload(*gin.Context, Filer, map[string][]*multipart.FileHeader) error
}
