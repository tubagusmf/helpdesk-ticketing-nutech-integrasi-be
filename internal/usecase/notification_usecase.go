package usecase

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type NotificationUsecase struct {
	notificationRepo model.INotificationRepository
}

func NewNotificationUsecase(
	notificationRepo model.INotificationRepository,
) model.INotificationUsecase {
	return &NotificationUsecase{
		notificationRepo: notificationRepo,
	}
}

func (u *NotificationUsecase) Create(
	ctx context.Context,
	in model.CreateNotificationInput,
) (*model.Notification, error) {

	notification := model.Notification{
		UserID:        in.UserID,
		ActorID:       in.ActorID,
		TicketID:      in.TicketID,
		Type:          in.Type,
		ReferenceType: in.ReferenceType,
		ReferenceID:   in.ReferenceID,
		Title:         in.Title,
		Message:       in.Message,
		IsRead:        false,
	}

	return u.notificationRepo.Create(ctx, notification)
}

func (u *NotificationUsecase) FindAllByUserID(
	ctx context.Context,
	userID int64,
) ([]*model.NotificationResponse, error) {
	return u.notificationRepo.FindAllByUserID(ctx, userID)
}

func (u *NotificationUsecase) MarkAsRead(
	ctx context.Context,
	id int64,
	userID int64,
) error {
	return u.notificationRepo.MarkAsRead(ctx, id, userID)
}

func (u *NotificationUsecase) CountUnread(
	ctx context.Context,
	userID int64,
) (int64, error) {
	return u.notificationRepo.CountUnread(ctx, userID)
}
