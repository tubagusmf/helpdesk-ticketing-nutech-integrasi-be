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
            u.name AS user_name,
            old_staff.name AS old_user_name,
            new_staff.name AS new_user_name
        `).
		Joins(`
            LEFT JOIN users u
                ON u.id = th.user_id
        `).
		Joins(`
            LEFT JOIN users old_staff
                ON old_staff.id = CASE
                    WHEN th.action = 'STAFF_REASSIGNED'
                     AND th.field_name = 'staff_assigned_to_id'
                    THEN NULLIF(th.old_value, '')::BIGINT
                    ELSE NULL
                END
        `).
		Joins(`
            LEFT JOIN users new_staff
                ON new_staff.id = CASE
                    WHEN th.action = 'STAFF_REASSIGNED'
                     AND th.field_name = 'staff_assigned_to_id'
                    THEN NULLIF(th.new_value, '')::BIGINT
                    ELSE NULL
                END
        `).
		Where("th.ticket_id = ?", ticketID).
		Order("th.created_at DESC, th.id DESC").
		Find(&histories)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return histories, nil
}
