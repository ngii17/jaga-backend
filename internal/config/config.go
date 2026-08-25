package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort               string
	DatabaseURL           string
	JWTSecret             string
	SMTPHost              string
	SMTPPort              string
	SMTPEmail             string
	SMTPAppPass           string
	EvidenceEncryptionKey string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: file .env tidak ditemukan, memakai environment variable sistem.")
	}

	return &Config{
		AppPort:               os.Getenv("APP_PORT"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		SMTPHost:              os.Getenv("SMTP_HOST"),
		SMTPPort:              os.Getenv("SMTP_PORT"),
		SMTPEmail:             os.Getenv("SMTP_EMAIL"),
		SMTPAppPass:           os.Getenv("SMTP_APP_PASSWORD"),
		EvidenceEncryptionKey: os.Getenv("EVIDENCE_ENCRYPTION_KEY"),
	}
}
