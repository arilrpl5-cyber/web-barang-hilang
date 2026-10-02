package models

import "time"

type ActivityLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      *uint     `gorm:"index" json:"user_id"`
	Action      string    `gorm:"size:100;index" json:"action"`
	Description string    `gorm:"type:text" json:"description"`
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
