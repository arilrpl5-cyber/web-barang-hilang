package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"lost-found-school/config"
	"lost-found-school/models"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// loginRedirect mengarahkan halaman ke login dengan info tujuan (next)
// agar setelah login user kembali ke halaman yang diminta.
func loginRedirect(c *gin.Context) {
	next := c.Request.URL.RequestURI()
	target := "/login"
	if next != "" && next != "/login" && next != "/register" {
		target = "/login?next=" + url.QueryEscape(next)
	}
	c.Redirect(http.StatusFound, target)
	c.Abort()
}

func AuthRequired(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""
		if ck, err := c.Cookie("token"); err == nil {
			tokenStr = ck
		}
		if tokenStr == "" {
			h := c.GetHeader("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				tokenStr = strings.TrimPrefix(h, "Bearer ")
			}
		}
		if tokenStr == "" {
			// Halaman HTML wajib login dulu -> arahkan ke /login (dengan next).
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				utils.Fail(c, http.StatusUnauthorized, "Unauthorized, silakan login")
				c.Abort()
				return
			}
			loginRedirect(c)
			return
		}
		claims, err := utils.ParseToken(tokenStr, cfg.JWTSecret)
		if err != nil {
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				utils.Fail(c, http.StatusUnauthorized, "Token tidak valid")
				c.Abort()
				return
			}
			loginRedirect(c)
			return
		}
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			utils.Fail(c, http.StatusUnauthorized, "User tidak ditemukan")
			c.Abort()
			return
		}
		if user.Status != "ACTIVE" {
			utils.Fail(c, http.StatusForbidden, "Akun tidak aktif")
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.Next()
	}
}
