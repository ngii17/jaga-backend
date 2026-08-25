package repository

import (
	"time"

	"jaga-backend/internal/models"

	"gorm.io/gorm"
)

type ThreatEntityRepository interface {
	SearchByName(name string, limit int) ([]models.ThreatEntity, error)
	FindExactByName(name string) (*models.ThreatEntity, error)
	IncrementOrCreate(name string, category models.ReportCategory) error
}

type threatEntityRepository struct {
	db *gorm.DB
}

func NewThreatEntityRepository(db *gorm.DB) ThreatEntityRepository {
	return &threatEntityRepository{db: db}
}

func (r *threatEntityRepository) SearchByName(name string, limit int) ([]models.ThreatEntity, error) {
	var results []models.ThreatEntity
	err := r.db.Raw(`
		SELECT *, similarity(name, ?) AS score
		FROM threat_entities
		WHERE name % ?
		ORDER BY score DESC
		LIMIT ?
	`, name, name, limit).Scan(&results).Error
	return results, err
}

func (r *threatEntityRepository) FindExactByName(name string) (*models.ThreatEntity, error) {
	var entity models.ThreatEntity
	err := r.db.Where("name = ?", name).First(&entity).Error
	return &entity, err
}

// IncrementOrCreate dipanggil saat sebuah laporan diverifikasi diterima (dipakai nanti di Modul 3).
func (r *threatEntityRepository) IncrementOrCreate(name string, category models.ReportCategory) error {
	var entity models.ThreatEntity
	err := r.db.Where("name = ?", name).First(&entity).Error

	now := time.Now()
	if err != nil {
		// Belum pernah ada -> buat baru
		newEntity := models.ThreatEntity{
			Name:            name,
			Category:        category,
			ReportCount:     1,
			FirstReportedAt: now,
			LastReportedAt:  now,
		}
		return r.db.Create(&newEntity).Error
	}

	// Sudah ada -> tambah hitungan
	entity.ReportCount++
	entity.LastReportedAt = now
	return r.db.Save(&entity).Error
}
