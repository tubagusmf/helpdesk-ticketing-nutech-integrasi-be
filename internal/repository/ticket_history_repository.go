package repository

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketHistoryRepo struct {
	db *gorm.DB
}

func NewTicketHistoryRepo(db *gorm.DB) model.ITicketHistoryRepository {
	return &TicketHistoryRepo{db: db}
}

func (r *TicketHistoryRepo) Create(ctx context.Context, history model.TicketHistory) (*model.TicketHistory, error) {
	if err := r.db.WithContext(ctx).Create(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *TicketHistoryRepo) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketHistory, error) {
	var histories []*model.TicketHistory

	err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		Find(&histories).Error

	if err != nil {
		return nil, err
	}

	return histories, nil
}
