package httpHelper

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func (i *HTTPHelper) HandleError(c *gin.Context, err error, code int) bool {
	if err != nil && err.Error() != sql.ErrNoRows.Error() {
		c.JSON(code, err.Error())
		return true
	}
	return false
}
