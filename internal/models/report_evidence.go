package models

import "time"

type ReportEvidence struct {
	ID               uint      `gorm:"primaryKey"`
	ReportID         uint      `gorm:"not null;index"`
	FileName         string    `gorm:"not null"`
	FilePath         string    `gorm:"not null"`
	DetectedMimeType string    `gorm:"not null"`
	FileSize         int64     `gorm:"not null"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

func (ReportEvidence) TableName() string {
	return "report_evidence"
}
