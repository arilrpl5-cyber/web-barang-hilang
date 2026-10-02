package models

import "time"

type Match struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	LostReportID  uint      `gorm:"index;not null" json:"lost_report_id"`
	FoundReportID uint      `gorm:"index;not null" json:"found_report_id"`
	Score         float64   `json:"score"`
	Status        string    `gorm:"size:20;default:PENDING" json:"status"` // PENDING, REVIEWING, CONFIRMED, REJECTED
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	LostReport  Report `gorm:"foreignKey:LostReportID" json:"lost_report,omitempty"`
	FoundReport Report `gorm:"foreignKey:FoundReportID" json:"found_report,omitempty"`
}
