package repository

import (
	"jaga-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepository interface {
	CreateFull(report *models.Report, contact *models.ReportContact, evidences []models.ReportEvidence) error
	FindByID(id uint) (*models.Report, error)
	FindAll(status models.ReportStatus) ([]models.Report, error)
	UpdateStatus(reportID uint, newStatus models.ReportStatus) error
	CreateStatusLog(log *models.ReportStatusLog) error
	FindEvidenceByID(evidenceID uint) (*models.ReportEvidence, error)
	FindEvidencesByReportID(reportID uint) ([]models.ReportEvidence, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// CreateFull menyimpan Report + ReportContact (opsional) + ReportEvidence sekaligus dalam satu transaction.
func (r *reportRepository) CreateFull(report *models.Report, contact *models.ReportContact, evidences []models.ReportEvidence) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		if contact != nil {
			contact.ReportID = report.ID
			if err := tx.Create(contact).Error; err != nil {
				return err
			}
		}

		for i := range evidences {
			evidences[i].ReportID = report.ID
			if err := tx.Create(&evidences[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reportRepository) FindByID(id uint) (*models.Report, error) {
	var report models.Report
	if err := r.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) FindAll(status models.ReportStatus) ([]models.Report, error) {
	var reports []models.Report
	query := r.db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at desc").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *reportRepository) UpdateStatus(reportID uint, newStatus models.ReportStatus) error {
	return r.db.Model(&models.Report{}).Where("id = ?", reportID).Update("status", newStatus).Error
}

func (r *reportRepository) CreateStatusLog(log *models.ReportStatusLog) error {
	return r.db.Create(log).Error
}

var _ = uuid.UUID{}

func (r *reportRepository) FindEvidenceByID(evidenceID uint) (*models.ReportEvidence, error) {
	var evidence models.ReportEvidence
	if err := r.db.First(&evidence, evidenceID).Error; err != nil {
		return nil, err
	}
	return &evidence, nil
}

func (r *reportRepository) FindEvidencesByReportID(reportID uint) ([]models.ReportEvidence, error) {
	var evidences []models.ReportEvidence
	if err := r.db.Where("report_id = ?", reportID).Find(&evidences).Error; err != nil {
		return nil, err
	}
	return evidences, nil
}
