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

func (r *TicketCommentRepo) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketComment, error) {
	var comments []*model.TicketComment

	err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	return comments, nil
}
