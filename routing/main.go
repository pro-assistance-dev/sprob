package routing

import (
	"reflect"

	"github.com/pro-assistance-dev/sprob/handlers/basehandler"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
)

type IHandler interface {
	GetAll(c *gin.Context)
	LabelValue(c *gin.Context)
	FTSP(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Delete(c *gin.Context)
	Update(c *gin.Context)
}

type RouterOpts struct {
	ws  *gin.RouterGroup
	key string
	h   IHandler
}

type option func(*RouterOpts)

func WithWS(ws *gin.RouterGroup) option {
	return func(opts *RouterOpts) {
		opts.ws = ws
	}
}

func WithHandler(h IHandler) option {
	return func(opts *RouterOpts) {
		opts.h = h
	}
}

func WithKey(key string) option {
	return func(opts *RouterOpts) {
		opts.key = key
	}
}

func handleOpts[T basehandler.Relationable](opts ...option) RouterOpts {
	routerOpts := RouterOpts{}
	for _, opt := range opts {
		opt(&routerOpts)
	}
	if routerOpts.key == "" {
		routerOpts.key = getKey[T]()
	}
	if routerOpts.h == nil {
		handler := basehandler.InitH[T]()
		routerOpts.h = &handler
	}
	return routerOpts
}

func getKey[T basehandler.Relationable]() string {
	key := reflect.TypeFor[T]().Name()
	pluralKey := pluralize.NewClient().Plural(key)
	kebab := strcase.ToKebab(pluralKey)
	return kebab
}

func InitR[T basehandler.Relationable](routerGroup *gin.RouterGroup, opts ...option) *gin.RouterGroup {
	routerOpts := handleOpts[T](opts...)
	r := routerGroup.Group(routerOpts.key)
	if routerOpts.ws != nil {
		routerOpts.ws = routerOpts.ws.Group(routerOpts.key)
	}

	r.GET("/label-value/:label/:value", routerOpts.h.LabelValue)
	r.GET("", routerOpts.h.GetAll)
	r.POST("/ftsp", routerOpts.h.FTSP)
	r.GET("/:id", routerOpts.h.Get)
	r.POST("", routerOpts.h.Create)
	r.DELETE("/:id", routerOpts.h.Delete)
	r.PUT("/:id", routerOpts.h.Update)
	return r
}
