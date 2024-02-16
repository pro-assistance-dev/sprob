package http

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/config"
)

type HTTP struct {
	HTTPS        bool
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewHTTP(config config.Server) *HTTP {
	return &HTTP{Host: config.Host, Port: config.Port, middleware: createMiddleware(), ReadTimeout: time.Duration(config.ReadTimeout), WriteTimeout: time.Duration(config.WriteTimeout), HTTPS: config.HTTPS}
}

func (i *HTTP) SetFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
}

func (i *HTTP) ListenAndServe(handler http.Handler) {
	srv := &http.Server{
		ReadTimeout:  i.ReadTimeout * time.Second,
		WriteTimeout: i.WriteTimeout * time.Second,
		Handler:      handler,
		Addr:         fmt.Sprintf(":%s", i.Port),
	}
	var err error
	if i.HTTPS {
		err = srv.ListenAndServeTLS("localhost.crt", "localhost.key")
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.Fatalln(err)
	}
}

// func (i *HTTP) CORSMiddleware() gin.HandlerFunc {
// 	// return i.middleware.corsMiddleware()
// }
