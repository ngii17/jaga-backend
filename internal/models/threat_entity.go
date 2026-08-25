package models

import "time"

type ThreatEntity struct {
	ID              uint           `gorm:"primaryKey"`
	Name            string         `gorm:"not null;index"`
	Category        ReportCategory `gorm:"type:varchar(30);not null"`
	ReportCount     int            `gorm:"default:0"`
	FirstReportedAt time.Time
	LastReportedAt  time.Time
}

func (ThreatEntity) TableName() string {
	return "threat_entities"
}
