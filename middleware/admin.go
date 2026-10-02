package middleware

import (
	"net/http"
	"strings"

	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		if role != "ADMIN" {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				utils.Fail(c, http.StatusForbidden, "Akses admin diperlukan")
				c.Abort()
				return
			}
			c.HTML(http.StatusForbidden, "403.html", gin.H{"title": "403"})
			c.Abort()
			return
		}
		c.Next()
	}
}
