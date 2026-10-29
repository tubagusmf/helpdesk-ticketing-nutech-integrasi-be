package repository

import (
	"context"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FCMTokenRepo struct {
	db *gorm.DB
}

func NewFCMTokenRepo(db *gorm.DB) model.IUserFCMTokenRepository {
	return &FCMTokenRepo{db: db}
}

func (r *FCMTokenRepo) UpsertToken(ctx context.Context, userID int64, token string) error {
	now := time.Now()

	device := model.UserFCMToken{
		UserID:           userID,
		Token:            token,
		IsActive:         true,
		LastRegisteredAt: now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "token"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"user_id":            userID,
				"is_active":          true,
				"last_registered_at": now,
				"updated_at":         now,
			}),
		}).
		Create(&device).Error
}

func (r *FCMTokenRepo) DeleteByUserIDAndToken(ctx context.Context, userID int64, token string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND token = ?", userID, token).
		Delete(&model.UserFCMToken{}).Error
}

func (r *FCMTokenRepo) DeleteToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&model.UserFCMToken{}).Error
}

func (r *FCMTokenRepo) FindActiveByUserID(ctx context.Context, userID int64) ([]model.UserFCMToken, error) {
	var devices []model.UserFCMToken

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_registered_at DESC").
		Find(&devices).Error

	return devices, err
}
