package httpHelper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/config"
)

type HTTPHelper struct {
	Host string
	Port string
}

func NewHTTPHelper(config config.Config) *HTTPHelper {
	return &HTTPHelper{Host: config.ServerHost, Port: config.ServerPort}
}

func (i *HTTPHelper) SetFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
}
