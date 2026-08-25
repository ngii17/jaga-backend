package repository

import (
	"jaga-backend/internal/models"

	"gorm.io/gorm"
)

type LegalEntityRepository interface {
	SearchByName(name string, limit int) ([]models.LegalEntity, error)
}

type legalEntityRepository struct {
	db *gorm.DB
}

func NewLegalEntityRepository(db *gorm.DB) LegalEntityRepository {
	return &legalEntityRepository{db: db}
}

// SearchByName mencari entitas legal berdasarkan kemiripan nama (fuzzy search),
// diurutkan dari yang paling mirip.
func (r *legalEntityRepository) SearchByName(name string, limit int) ([]models.LegalEntity, error) {
	var results []models.LegalEntity
	err := r.db.Raw(`
		SELECT *, similarity(official_name, ?) AS score
		FROM legal_entities
		WHERE official_name % ?
		ORDER BY score DESC
		LIMIT ?
	`, name, name, limit).Scan(&results).Error
	return results, err
}
