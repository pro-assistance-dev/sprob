package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pro-assistance/pro-assister/helper"
	"github.com/pro-assistance/pro-assister/helpers/sql"
)

type Middleware struct {
	helper *helper.Helper
}

func CreateMiddleware(helper *helper.Helper) *Middleware {
	return &Middleware{helper: helper}
}

func (m *Middleware) injectFTSP() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.URL.Path, "ftsp") {
			return
		}
		ftspQuery := &sql.FTSPQuery{}
		err := ftspQuery.FromForm(c)
		if m.helper.HTTP.HandleError(c, err) {
			return
		}

		ftsp, found := ftspStore.GetOrCreateFTSP(ftspQuery)

		if !found {
			c.JSON(http.StatusOK, nil)
			c.Abort()
			return
		}

		m.helper.SQL.InjectFTSP2(c.Request, &ftsp)

		if err != nil {
			return
		}
		c.Next()
	}
}

func (m *Middleware) InjectRequestInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.injectClaims(c)
		m.injectFTSP()
		c.Next()
	}
}

func (m *Middleware) injectClaims(c *gin.Context) {
	err := Claims{ClaimUserID, ClaimDomainIDS}.Inject(c.Request, m.helper.Token)
	if m.helper.HTTP.HandleError(c, err) {
		return
	}
	if err != nil {
		return
	}
	if err != nil {
		return
	}
}

func (m *Middleware) methodIsAllowed(requestMethod string) bool {
	allowedMethods := []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"}
	for _, allowedMethod := range allowedMethods {
		if requestMethod == allowedMethod {
			return true
		}
	}
	return false
}

func (m *Middleware) CORSMiddleware() gin.HandlerFunc {
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

func (m *Middleware) CheckPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if !m.checkPermission(c) {
		//	c.AbortWithStatus(http.StatusForbidden)
		//	return
		//}
		c.Next()
	}
}
