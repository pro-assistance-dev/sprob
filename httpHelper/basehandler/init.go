package basehandler

import (
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
	Create(*gin.Context, *TSingle) error
	GetAll(*gin.Context) (TPluralWithCount, error)
	Get(*gin.Context, string) (*TSingle, error)
	Delete(*gin.Context, string) error
	Update(*gin.Context, *TSingle) error
}

type IServiceWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IService[TSingle, TPlural, TPluralWithCount]
	UpsertMany(*TPlural) error
	DeleteMany([]uuid.UUID) error
}

type IRepository[TSingle, TPlural, TPluralWithCount any] interface {
	Create(*gin.Context, *TSingle) error
	Update(*gin.Context, *TSingle) error
	GetAll(*gin.Context) (TPluralWithCount, error)
	Get(*gin.Context, string) (*TSingle, error)
	Delete(*gin.Context, string) error
}

type IRepositoryWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IRepository[TSingle, TPlural, TPluralWithCount]
	Upsert(*gin.Context, *TSingle) error
	UpsertMany(*gin.Context, *TPlural) error
	DeleteMany(*gin.Context, []uuid.UUID) error
}

type IFilesService interface {
	Upload(*gin.Context, Filer, map[string][]*multipart.FileHeader) error
}
