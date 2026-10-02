package controllers

import (
	"net/http"

	"lost-found-school/models"
	"lost-found-school/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationController struct{ DB *gorm.DB }

func (h *NotificationController) List(c *gin.Context) {
	u := currentUser(c)
	var notifs []models.Notification
	h.DB.Where("user_id = ?", u.ID).Order("created_at DESC").Limit(50).Find(&notifs)
	var unread int64
	h.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", u.ID, false).Count(&unread)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Data notifikasi", "data": notifs, "unread": unread})
}

func (h *NotificationController) UnreadCount(c *gin.Context) {
	u := currentUser(c)
	var unread int64
	h.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", u.ID, false).Count(&unread)
	utils.Success(c, http.StatusOK, "Unread count", gin.H{"unread": unread})
}

func (h *NotificationController) MarkRead(c *gin.Context) {
	u := currentUser(c)
	h.DB.Model(&models.Notification{}).Where("id = ? AND user_id = ?", c.Param("id"), u.ID).Update("is_read", true)
	utils.Success(c, http.StatusOK, "Notifikasi ditandai dibaca", nil)
}

func (h *NotificationController) ReadAll(c *gin.Context) {
	u := currentUser(c)
	h.DB.Model(&models.Notification{}).Where("user_id = ?", u.ID).Update("is_read", true)
	utils.Success(c, http.StatusOK, "Semua notifikasi ditandai dibaca", nil)
}
