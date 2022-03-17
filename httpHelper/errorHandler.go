package httpHelper

import (
	"github.com/gin-gonic/gin"
)

func (i *HTTPHelper) HandleError(c *gin.Context, err error, code int) bool {
	if err != nil {
		c.JSON(code, err.Error())
		return true
	}
	return false
}
