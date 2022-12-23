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
	Host         string
	Port         string
	middleware   *middleware
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewHTTPHelper(config config.Server) *HTTPHelper {
	return &HTTPHelper{Host: config.Host, Port: config.Port, middleware: createMiddleware(), ReadTimeout: time.Duration(config.ReadTimeout), WriteTimeout: time.Duration(config.WriteTimeout)}
}

func (i *HTTPHelper) SetFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
}

func (i *HTTPHelper) ListenAndServe(handler http.Handler) {
	srv := &http.Server{
		ReadTimeout:  i.ReadTimeout * time.Second,
		WriteTimeout: i.WriteTimeout * time.Second,
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
