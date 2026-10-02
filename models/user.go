package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	NisNisn      string    `gorm:"size:50;uniqueIndex;not null" json:"nis_nisn"`
	Email        string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:20;default:USER" json:"role"` // USER, ADMIN
	ProfilePhoto string    `gorm:"size:255" json:"profile_photo"`
	Status       string    `gorm:"size:20;default:ACTIVE" json:"status"` // ACTIVE, INACTIVE, BANNED
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Reports []Report `gorm:"foreignKey:UserID" json:"-"`
}
