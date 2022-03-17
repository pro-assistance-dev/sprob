package httpHelper

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type HTTPHelper struct {
}

func NewHTTPHelper() *HTTPHelper {
	return &HTTPHelper{}
}

func (i *HTTPHelper) SetFileHeaders(c *gin.Context, fileName string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
}
