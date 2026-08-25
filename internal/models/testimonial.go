package models

import "time"

type TestimonialStatus string

const (
	TestimonialPending   TestimonialStatus = "pending"
	TestimonialDisetujui TestimonialStatus = "disetujui"
	TestimonialDitolak   TestimonialStatus = "ditolak"
)

type Testimonial struct {
	ID                   uint              `gorm:"primaryKey"`
	ReportID             *uint             `gorm:"uniqueIndex"`
	Story                string            `gorm:"type:text;not null"`
	Status               TestimonialStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	SubmitterFingerprint string            `gorm:"index"`
	Upvotes              int               `gorm:"default:0"`
	CreatedAt            time.Time         `gorm:"autoCreateTime"`
}

func (Testimonial) TableName() string {
	return "testimonials"
}
