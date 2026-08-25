package models

type ReportContact struct {
	ID       uint `gorm:"primaryKey"`
	ReportID uint `gorm:"not null;uniqueIndex"`
	Name     string
	Email    string
	Phone    string
}

func (ReportContact) TableName() string {
	return "report_contacts"
}
