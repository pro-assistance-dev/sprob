package httpHelper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/config"
	"log"
	"net/http"
)

type HTTPHelper struct {
	Host       string
	Port       string
	middleware *middleware
}

func NewHTTPHelper(config config.Config) *HTTPHelper {
	return &HTTPHelper{Host: config.ServerHost, Port: config.ServerPort, middleware: createMiddleware()}
}

func (i *HTTPHelper) SetFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
}

func (i *HTTPHelper) ListenAndServe(handler http.Handler) {
	err := http.ListenAndServe(fmt.Sprintf(":%s", i.Port), handler)
	if err != nil {
		log.Fatalln(err)
	}
}

func (i *HTTPHelper) CORSMiddleware() gin.HandlerFunc {
	return i.middleware.corsMiddleware()
}
