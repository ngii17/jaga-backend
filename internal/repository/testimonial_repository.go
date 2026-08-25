package repository

import (
	"jaga-backend/internal/models"

	"gorm.io/gorm"
)

type FingerprintStats struct {
	TotalSubmitted int
	TotalDitolak   int
}

type TestimonialRepository interface {
	Create(testimonial *models.Testimonial) error
	FindApproved() ([]models.Testimonial, error)
	FindPending() ([]models.Testimonial, error)
	FindByID(id uint) (*models.Testimonial, error)
	UpdateStatus(id uint, status models.TestimonialStatus, story string) error
	AddUpvote(id uint) error
	GetFingerprintStats(fingerprint string) (*FingerprintStats, error)
	FindApprovedReportsWithoutTestimonial() ([]models.Report, error)
}

type testimonialRepository struct {
	db *gorm.DB
}

func NewTestimonialRepository(db *gorm.DB) TestimonialRepository {
	return &testimonialRepository{db: db}
}

func (r *testimonialRepository) Create(testimonial *models.Testimonial) error {
	return r.db.Create(testimonial).Error
}

func (r *testimonialRepository) FindApproved() ([]models.Testimonial, error) {
	var testimonials []models.Testimonial
	err := r.db.Where("status = ?", models.TestimonialDisetujui).
		Order("upvotes desc, created_at desc").
		Find(&testimonials).Error
	return testimonials, err
}

func (r *testimonialRepository) FindPending() ([]models.Testimonial, error) {
	var testimonials []models.Testimonial
	err := r.db.Where("status = ?", models.TestimonialPending).
		Order("created_at asc").
		Find(&testimonials).Error
	return testimonials, err
}

func (r *testimonialRepository) FindByID(id uint) (*models.Testimonial, error) {
	var testimonial models.Testimonial
	if err := r.db.First(&testimonial, id).Error; err != nil {
		return nil, err
	}
	return &testimonial, nil
}

func (r *testimonialRepository) UpdateStatus(id uint, status models.TestimonialStatus, story string) error {
	updates := map[string]interface{}{"status": status}
	if story != "" {
		updates["story"] = story
	}
	return r.db.Model(&models.Testimonial{}).Where("id = ?", id).Updates(updates).Error
}

func (r *testimonialRepository) AddUpvote(id uint) error {
	return r.db.Model(&models.Testimonial{}).Where("id = ?", id).
		UpdateColumn("upvotes", gorm.Expr("upvotes + 1")).Error
}

func (r *testimonialRepository) GetFingerprintStats(fingerprint string) (*FingerprintStats, error) {
	var total int64
	var ditolak int64

	if err := r.db.Model(&models.Testimonial{}).
		Where("submitter_fingerprint = ?", fingerprint).
		Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.Testimonial{}).
		Where("submitter_fingerprint = ? AND status = ?", fingerprint, models.TestimonialDitolak).
		Count(&ditolak).Error; err != nil {
		return nil, err
	}

	return &FingerprintStats{TotalSubmitted: int(total), TotalDitolak: int(ditolak)}, nil
}

func (r *testimonialRepository) FindApprovedReportsWithoutTestimonial() ([]models.Report, error) {
	var reports []models.Report
	err := r.db.Where("status = ? AND id NOT IN (SELECT report_id FROM testimonials WHERE report_id IS NOT NULL)",
		models.ReportStatusDiterima).
		Find(&reports).Error
	return reports, err
}
