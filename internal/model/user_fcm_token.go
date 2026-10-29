package model

import (
	"context"
	"time"
)

type UserFCMToken struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	UserID           int64     `json:"user_id" gorm:"not null;index"`
	Token            string    `json:"-" gorm:"type:text;not null;uniqueIndex"`
	IsActive         bool      `json:"is_active" gorm:"not null;default:true"`
	LastRegisteredAt time.Time `json:"last_registered_at" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (UserFCMToken) TableName() string {
	return "user_fcm_tokens"
}

type IUserFCMTokenRepository interface {
	UpsertToken(ctx context.Context, userID int64, token string) error
	DeleteByUserIDAndToken(ctx context.Context, userID int64, token string) error
	DeleteToken(ctx context.Context, token string) error
	FindActiveByUserID(ctx context.Context, userID int64) ([]UserFCMToken, error)
}

type IUserFCMTokenUsecase interface {
	RegisterToken(ctx context.Context, userID int64, token string) error
	UnregisterToken(ctx context.Context, userID int64, token string) error
	DeleteInvalidToken(ctx context.Context, token string) error
	FindActiveByUserID(ctx context.Context, userID int64) ([]UserFCMToken, error)
}
