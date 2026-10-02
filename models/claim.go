package models

import "time"

type Claim struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ReportID    uint       `gorm:"index;not null" json:"report_id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Evidence    string     `gorm:"type:text" json:"evidence"`
	Status      string     `gorm:"size:20;default:PENDING" json:"status"` // PENDING, REVIEWING, APPROVED, REJECTED, CANCELLED
	AdminNote   string     `gorm:"type:text" json:"admin_note"`
	ReviewedBy  *uint      `json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Report Report `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
