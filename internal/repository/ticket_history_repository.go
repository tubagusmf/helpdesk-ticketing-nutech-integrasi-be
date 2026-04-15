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

func (r *TicketHistoryRepo) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketHistoryResponse, error) {
	var histories []*model.TicketHistoryResponse

	tx := r.db.WithContext(ctx).
		Table("ticket_histories th").
		Select(`
			th.*,
			u.name as user_name
		`).
		Joins("LEFT JOIN users u ON u.id = th.user_id").
		Where("th.ticket_id = ?", ticketID).
		Order("th.created_at DESC").
		Find(&histories)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return histories, nil
}
