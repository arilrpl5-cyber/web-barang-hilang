package models

import "time"

type Report struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index;not null" json:"user_id"`
	CategoryID      uint      `gorm:"index" json:"category_id"`
	Type            string    `gorm:"size:10;index;not null" json:"type"` // LOST, FOUND
	ItemName        string    `gorm:"size:150;not null" json:"item_name"`
	Description     string    `gorm:"type:text" json:"description"`
	Photo           string    `gorm:"size:255" json:"photo"`
	Brand           string    `gorm:"size:100" json:"brand"`
	Color           string    `gorm:"size:50" json:"color"`
	Characteristics string    `gorm:"type:text" json:"characteristics"`
	Location        string    `gorm:"size:150" json:"location"`
	EventDate       string    `gorm:"size:20" json:"event_date"`
	EventTime       string    `gorm:"size:20" json:"event_time"`
	StorageLocation string    `gorm:"size:150" json:"storage_location"`
	AdditionalNote  string    `gorm:"type:text" json:"additional_note"`
	Status          string    `gorm:"size:20;index;default:PENDING" json:"status"` // PENDING, VERIFIED, PROCESSING, MATCHED, RETURNED, REJECTED, CANCELLED
	RejectReason    string    `gorm:"type:text" json:"reject_reason"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}
