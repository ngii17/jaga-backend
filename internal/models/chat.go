package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Fingerprint string    `gorm:"not null;index"`
	UsageCount  int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

type ChatMessageRole string

const (
	ChatRoleUser      ChatMessageRole = "user"
	ChatRoleAssistant ChatMessageRole = "assistant"
)

func (r ChatMessageRole) IsValid() bool {
	switch r {
	case ChatRoleUser, ChatRoleAssistant:
		return true
	}
	return false
}

type ChatMessage struct {
	ID        uint            `gorm:"primaryKey"`
	SessionID uuid.UUID       `gorm:"type:uuid;not null;index"`
	Role      ChatMessageRole `gorm:"type:varchar(20);not null"`
	Content   string          `gorm:"type:text;not null"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
