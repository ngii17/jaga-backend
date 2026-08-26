package models

import "time"

type ReportSeverity string

const (
	SeverityRingan ReportSeverity = "ringan"
	SeveritySedang ReportSeverity = "sedang"
	SeverityBerat  ReportSeverity = "berat"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusDiterima ReportStatus = "diterima"
	ReportStatusDitolak  ReportStatus = "ditolak"
)

type Report struct {
	ID                   uint           `gorm:"primaryKey"`
	Category             ReportCategory `gorm:"type:varchar(30);not null"`
	ReportedName         string         `gorm:"not null;index"`
	ReportedAccountOrWA  string
	Description          string         `gorm:"type:text;not null"`
	Severity             ReportSeverity `gorm:"type:varchar(20);not null"`
	IsAnonymous          bool           `gorm:"not null;default:true"`
	SubmitterFingerprint string         `gorm:"not null;index"`
	Status               ReportStatus   `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt            time.Time      `gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime"`
}

func (Report) TableName() string {
	return "reports"
}

func (c ReportCategory) IsValid() bool {
	switch c {
	case CategoryPinjolIlegal, CategoryJudiOnline, CategoryInvestasiBodong, CategoryReportLainnya:
		return true
	}
	return false
}

func (s ReportSeverity) IsValid() bool {
	switch s {
	case SeverityRingan, SeveritySedang, SeverityBerat:
		return true
	}
	return false
}

func (s ReportStatus) IsValid() bool {
	switch s {
	case ReportStatusPending, ReportStatusDiterima, ReportStatusDitolak:
		return true
	}
	return false
}
