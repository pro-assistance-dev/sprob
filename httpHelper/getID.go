package httpHelper

import "github.com/gin-gonic/gin"

func GetID(c *gin.Context) string {
	return c.Param("id")
}
