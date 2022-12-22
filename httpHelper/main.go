package httpHelper

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/config"
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
	srv := &http.Server{
		ReadTimeout:  1500 * time.Second,
		WriteTimeout: 1500 * time.Second,
		Handler:      handler,
		Addr:         fmt.Sprintf(":%s", i.Port),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}

func (i *HTTPHelper) CORSMiddleware() gin.HandlerFunc {
	return i.middleware.corsMiddleware()
}
