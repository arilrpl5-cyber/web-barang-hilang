package utils

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, message string, data interface{}) {
	if data == nil {
		c.JSON(status, gin.H{"success": true, "message": message})
		return
	}
	c.JSON(status, gin.H{"success": true, "message": message, "data": data})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "message": message})
}

func Pagination(c *gin.Context, defaultLimit int) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = 100
	}
	offset = (page - 1) * limit
	return
}

func Paginated(c *gin.Context, message string, items interface{}, total int64, page, limit int) {
	pages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    items,
		"meta": gin.H{
			"page": page, "limit": limit, "total": total, "pages": pages,
		},
	})
}
