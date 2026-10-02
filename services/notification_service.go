package services

import (
	"lost-found-school/models"

	"gorm.io/gorm"
)

func Notify(db *gorm.DB, userID uint, title, message, notifType string) {
	db.Create(&models.Notification{
		UserID: userID, Title: title, Message: message, Type: notifType,
	})
}

func LogActivity(db *gorm.DB, userID *uint, action, desc, ip string) {
	db.Create(&models.ActivityLog{
		UserID: userID, Action: action, Description: desc, IPAddress: ip,
	})
}
