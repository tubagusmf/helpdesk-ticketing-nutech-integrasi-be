package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type FCMTokenUsecase struct {
	repo model.IUserFCMTokenRepository
}

func NewFCMTokenUsecase(repo model.IUserFCMTokenRepository) model.IUserFCMTokenUsecase {
	return &FCMTokenUsecase{repo: repo}
}

func (u *FCMTokenUsecase) RegisterToken(ctx context.Context, userID int64, token string) error {
	token = strings.TrimSpace(token)

	if userID <= 0 {
		return errors.New("user ID tidak valid")
	}

	if token == "" || len(token) > 4096 {
		return errors.New("FCM token tidak valid")
	}

	return u.repo.UpsertToken(ctx, userID, token)
}

func (u *FCMTokenUsecase) UnregisterToken(ctx context.Context, userID int64, token string) error {
	token = strings.TrimSpace(token)

	if userID <= 0 || token == "" || len(token) > 4096 {
		return errors.New("data penghapusan FCM token tidak valid")
	}

	return u.repo.DeleteByUserIDAndToken(ctx, userID, token)
}

func (u *FCMTokenUsecase) DeleteInvalidToken(ctx context.Context, token string) error {
	return u.repo.DeleteToken(ctx, token)
}

func (u *FCMTokenUsecase) FindActiveByUserID(ctx context.Context, userID int64) ([]model.UserFCMToken, error) {
	return u.repo.FindActiveByUserID(ctx, userID)
}
