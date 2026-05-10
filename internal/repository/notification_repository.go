package repository

import (
	"context"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) model.INotificationRepository {
	return &NotificationRepo{
		db: db,
	}
}

func (r *NotificationRepo) Create(ctx context.Context, notification model.Notification) (*model.Notification, error) {
	notification.CreatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&notification).Error; err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationRepo) FindAllByUserID(ctx context.Context, userID int64) ([]*model.NotificationResponse, error) {
	var notifications []*model.NotificationResponse

	err := r.db.WithContext(ctx).
		Table("notifications").
		Joins("JOIN tickets ON tickets.id = notifications.ticket_id").
		Joins("JOIN users actor ON actor.id = notifications.actor_id").
		Where("notifications.user_id = ?", userID).
		Select(`
			notifications.id,
			notifications.ticket_id,
			tickets.ticket_code,
			actor.name as actor_name,
			notifications.type,
			notifications.title,
			notifications.message,
			notifications.is_read,
			notifications.created_at
		`).
		Order("notifications.created_at DESC").
		Scan(&notifications).Error

	return notifications, err
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *NotificationRepo) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&total).Error

	return total, err
}

func (r *NotificationRepo) Delete(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Notification{}).Error
}

func (r *NotificationRepo) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&model.Notification{}).Error
}
