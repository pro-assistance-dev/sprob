package http

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func (i *HTTP) HandleError(c *gin.Context, err error) bool {
	if err != nil && err.Error() != sql.ErrNoRows.Error() {
		_ = c.Error(err)
		// 401 для проблем с токеном, 500 для остального (см. authErrors.go).
		// ⚠️ Раньше здесь было ТОЧНОЕ сравнение строки `"Token is expired"` —
		// из-за него невалидная подпись и «expired by 1h2m» отдавались как 500.
		c.JSON(StatusForError(err), err.Error())
		return true
	}
	return false
}
