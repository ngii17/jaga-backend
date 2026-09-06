package repository

import (
	"time"

	"jaga-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	FindSessionByFingerprint(fingerprint string) (*models.ChatSession, error)
	CreateSession(fingerprint string) (*models.ChatSession, error)
	IncrementUsage(sessionID uuid.UUID) error
	SaveMessage(sessionID uuid.UUID, role models.ChatMessageRole, content string) error
	GetMessagesBySession(sessionID uuid.UUID) ([]models.ChatMessage, error)
	DeleteOldSessions(olderThan time.Time) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) FindSessionByFingerprint(fingerprint string) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.db.Where("fingerprint = ?", fingerprint).First(&session).Error
	return &session, err
}

func (r *chatRepository) CreateSession(fingerprint string) (*models.ChatSession, error) {
	session := &models.ChatSession{
		Fingerprint: fingerprint,
		UsageCount:  0,
	}
	err := r.db.Create(session).Error
	return session, err
}

func (r *chatRepository) IncrementUsage(sessionID uuid.UUID) error {
	return r.db.Model(&models.ChatSession{}).
		Where("id = ?", sessionID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *chatRepository) SaveMessage(sessionID uuid.UUID, role models.ChatMessageRole, content string) error {
	message := &models.ChatMessage{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	}
	return r.db.Create(message).Error
}

func (r *chatRepository) GetMessagesBySession(sessionID uuid.UUID) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

// DeleteOldSessions menghapus sesi (beserta pesannya) yang lebih tua dari waktu tertentu.
// Dipanggil oleh "petugas kebersihan" (goroutine+ticker) di main.go.
func (r *chatRepository) DeleteOldSessions(olderThan time.Time) error {
	var oldSessions []models.ChatSession
	if err := r.db.Where("created_at < ?", olderThan).Find(&oldSessions).Error; err != nil {
		return err
	}

	for _, session := range oldSessions {
		if err := r.db.Where("session_id = ?", session.ID).Delete(&models.ChatMessage{}).Error; err != nil {
			return err
		}
	}

	return r.db.Where("created_at < ?", olderThan).Delete(&models.ChatSession{}).Error
}
