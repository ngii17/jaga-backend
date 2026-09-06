package database

import (
	"log"

	"jaga-backend/internal/config"
	"jaga-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal konek ke database: %v", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		log.Fatalf("Gagal mengaktifkan ekstensi pg_trgm: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.LegalEntity{},
		&models.ThreatEntity{},
		&models.QuizQuestion{},
		&models.RiskLevel{},
		&models.Report{},
		&models.ReportContact{},
		&models.ReportEvidence{},
		&models.ReportStatusLog{},
		&models.Testimonial{},
		&models.ChatSession{},
		&models.ChatMessage{},
	); err != nil {
		log.Fatalf("Gagal migrasi tabel: %v", err)
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_legal_entities_name_trgm ON legal_entities USING GIN (official_name gin_trgm_ops)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threat_entities_name_trgm ON threat_entities USING GIN (name gin_trgm_ops)")

	log.Println("Database terkoneksi & migrasi users selesai.")
	return db
}
