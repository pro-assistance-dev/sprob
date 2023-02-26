package basehandler

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type IHandler interface {
	GetAll(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Delete(c *gin.Context)
	Update(c *gin.Context)
}

type IService[TSingle, TPlural, TPluralWithCount any] interface {
	SetQueryFilter(*gin.Context) error

	Create(*TSingle) error
	GetAll() (TPluralWithCount, error)
	Get(string) (*TSingle, error)
	Delete(string) error
	Update(*TSingle) error
}

type IRepository[TSingle, TPlural, TPluralWithCount any] interface {
	SetQueryFilter(*gin.Context) error
	DB() *bun.DB

	Create(*TSingle) error
	GetAll() (TPluralWithCount, error)
	Get(string) (*TSingle, error)
	Delete(string) error
	Update(*TSingle) error
}

type IFilesService interface {
	Upload(*gin.Context, Filer, map[string][]*multipart.FileHeader) error
}
