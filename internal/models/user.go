package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin       UserRole = "admin"
	RoleVerifikator UserRole = "verifikator"
)

type User struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name             string    `gorm:"not null"`
	Email            string    `gorm:"not null;unique"`
	PasswordHash     string    `gorm:"not null"`
	Role             UserRole  `gorm:"type:varchar(20);not null"`
	CanApproveReport bool      `gorm:"default:false"`
	CanDeleteReport  bool      `gorm:"default:false"`
	CanManageUsers   bool      `gorm:"default:false"`
	OTPCodeHash      *string
	OTPExpiresAt     *time.Time
	OTPAttempts      int        `gorm:"default:0"`
	CreatedBy        *uuid.UUID `gorm:"type:uuid"`
	CreatedAt        time.Time  `gorm:"autoCreateTime"`
}

func (User) TableName() string {
	return "users"
}
