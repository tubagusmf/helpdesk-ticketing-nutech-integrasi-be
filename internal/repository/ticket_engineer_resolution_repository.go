package repository

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketEngineerResolutionRepo struct {
	db *gorm.DB
}

func NewTicketEngineerResolutionRepo(
	db *gorm.DB,
) model.ITicketEngineerResolutionRepository {
	return &TicketEngineerResolutionRepo{
		db: db,
	}
}

func (r *TicketEngineerResolutionRepo) Create(ctx context.Context, resolution *model.TicketEngineerResolution) error {
	return r.db.WithContext(ctx).Create(resolution).Error
}

func (r *TicketEngineerResolutionRepo) CreateAttachment(ctx context.Context, attachment *model.TicketEngineerResolutionAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

func (r *TicketEngineerResolutionRepo) FindByTicketID(ctx context.Context, ticketID int64) (*model.TicketEngineerResolution, error) {
	var resolution model.TicketEngineerResolution

	err := r.db.WithContext(ctx).
		Preload("Attachments").
		Where("ticket_id = ?", ticketID).
		First(&resolution).
		Error

	if err != nil {
		return nil, err
	}

	return &resolution, nil
}
