package models

import (
	"time"

	"github.com/google/uuid"
)

type ReportStatusLog struct {
	ID        uint         `gorm:"primaryKey"`
	ReportID  uint         `gorm:"not null;index"`
	OldStatus ReportStatus `gorm:"type:varchar(20);not null"`
	NewStatus ReportStatus `gorm:"type:varchar(20);not null"`
	ChangedBy uuid.UUID    `gorm:"type:uuid;not null"`
	Note      string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (ReportStatusLog) TableName() string {
	return "report_status_logs"
}
