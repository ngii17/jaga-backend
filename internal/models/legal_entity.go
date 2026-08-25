package models

import "time"

type LegalCategory string

const (
	CategoryPinjol      LegalCategory = "pinjol"
	CategoryInvestasi   LegalCategory = "investasi"
	CategoryJudiTerkait LegalCategory = "judi_terkait"
	CategoryLainnya     LegalCategory = "lainnya"
)

type LegalEntity struct {
	ID               uint          `gorm:"primaryKey"`
	OfficialName     string        `gorm:"not null;index"`
	Category         LegalCategory `gorm:"type:varchar(30);not null"`
	OJKLicenseNumber string
	Status           string    `gorm:"type:varchar(20);not null;default:'legal'"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

func (LegalEntity) TableName() string {
	return "legal_entities"
}
