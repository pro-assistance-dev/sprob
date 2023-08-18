package httpHelper

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *HTTPHelper) HandleError(c *gin.Context, err error) bool {
	if err != nil && err.Error() != sql.ErrNoRows.Error() {
		_ = c.Error(err)
		code := http.StatusInternalServerError
		if err.Error() == "Token is expired" {
			code = http.StatusUnauthorized
		}
		c.JSON(code, err.Error())
		return true
	}
	return false
}
