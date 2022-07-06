package httpHelper

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type middleware struct {
}

func createMiddleware() *middleware {
	return &middleware{}
}

func (m *middleware) methodIsAllowed(requestMethod string) bool {
	allowedMethods := []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"}
	for _, allowedMethod := range allowedMethods {
		if requestMethod == allowedMethod {
			return true
		}
	}
	return false
}

func (m *middleware) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		if !m.methodIsAllowed(c.Request.Method) {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		c.Next()
	}
}
