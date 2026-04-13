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
	if err := r.db.WithContext(ctx).Create(&comment).Error; err != nil {
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
