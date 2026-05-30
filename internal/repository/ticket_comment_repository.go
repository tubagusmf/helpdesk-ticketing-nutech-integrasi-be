package repository

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketCommentRepo struct {
	db *gorm.DB
}

func NewTicketCommentRepo(db *gorm.DB) model.ITicketCommentRepository {
	return &TicketCommentRepo{db: db}
}

func (r *TicketCommentRepo) Create(ctx context.Context, comment model.TicketComment) (*model.TicketComment, error) {
	if err := r.db.WithContext(ctx).
		Create(&comment).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&comment, comment.ID).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *TicketCommentRepo) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketCommentResponse, error) {
	var comments []*model.TicketCommentResponse

	err := r.db.WithContext(ctx).
		Table("ticket_comments").
		Select(`
			ticket_comments.id,
			ticket_comments.ticket_id,
			ticket_comments.user_id,
			users.name as user_name,
			ticket_comments.message,
			ticket_comments.created_at
		`).
		Joins("LEFT JOIN users ON users.id = ticket_comments.user_id").
		Where("ticket_comments.ticket_id = ?", ticketID).
		Order("ticket_comments.created_at ASC").
		Scan(&comments).Error

	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *TicketCommentRepo) CountUnreadByTicket(ctx context.Context, ticketID int64, role string, userID int64) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Table("ticket_comments").
		Where("ticket_id = ?", ticketID).
		Where("user_id != ?", userID)

	if role == "USER" {

		query = query.Where("is_read_by_user = false")

	} else if role == "STAFF" {

		query = query.Where("is_read_by_staff = false")

	} else if role == "ADMINISTRATOR" {

		query = query.Where("is_read_by_administrator = false")
	}

	err := query.Count(&count).Error

	return count, err
}

func (r *TicketCommentRepo) MarkAsRead(ctx context.Context, ticketID int64, role string) error {
	updates := map[string]interface{}{}

	if role == "USER" {

		updates["is_read_by_user"] = true

	} else if role == "STAFF" {

		updates["is_read_by_staff"] = true

	} else if role == "ADMINISTRATOR" {

		updates["is_read_by_administrator"] = true
	}

	return r.db.WithContext(ctx).
		Table("ticket_comments").
		Where("ticket_id = ?", ticketID).
		Updates(updates).Error
}
