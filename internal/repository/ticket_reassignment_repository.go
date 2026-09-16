package repository

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketReassignmentRepo struct {
	db *gorm.DB
}

func NewTicketReassignmentRepo(db *gorm.DB) model.ITicketReassignmentRepository {
	return &TicketReassignmentRepo{
		db: db,
	}
}

func (r *TicketReassignmentRepo) Create(ctx context.Context, reassignment *model.TicketReassignment) error {
	return r.db.WithContext(ctx).Create(reassignment).Error
}

func (r *TicketReassignmentRepo) CreateAttachment(ctx context.Context, attachment *model.TicketReassignmentAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

func (r *TicketReassignmentRepo) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketReassignment, error) {
	var data []*model.TicketReassignment

	err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Preload("Attachments").
		Order("created_at DESC").
		Find(&data).
		Error

	return data, err
}
