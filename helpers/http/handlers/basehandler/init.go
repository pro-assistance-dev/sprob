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
	Create(*TSingle) error
	GetAll() (TPluralWithCount, error)
	Get(string) (*TSingle, error)
	Delete(string) error
	Update(*TSingle) error
	SetQueryFilter(c *gin.Context) error
}

type IServiceWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IService[TSingle, TPlural, TPluralWithCount]
	UpsertMany(TPlural) error
	DeleteMany([]uuid.UUID) error
}

type IRepository[TSingle, TPlural, TPluralWithCount any] interface {
	Create(*TSingle) error
	Update(*TSingle) error
	GetAll() (TPluralWithCount, error)
	Get(string) (*TSingle, error)
	Delete(string) error

	SetQueryFilter(c *gin.Context) error
}

type IRepositoryWithMany[TSingle, TPlural, TPluralWithCount any] interface {
	IRepository[TSingle, TPlural, TPluralWithCount]
	Upsert(*TSingle) error
	UpsertMany(TPlural) error
	DeleteMany([]uuid.UUID) error
}

type IFilesService interface {
	Upload(*gin.Context, Filer, map[string][]*multipart.FileHeader) error
}
