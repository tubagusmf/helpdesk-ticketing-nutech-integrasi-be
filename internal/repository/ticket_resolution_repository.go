package repository

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketResolutionRepo struct {
	db *gorm.DB
}

func NewTicketResolutionRepo(db *gorm.DB) model.ITicketResolutionRepository {
	return &TicketResolutionRepo{db: db}
}

func (r *TicketResolutionRepo) Create(ctx context.Context, tx interface{}, resolution model.TicketResolution) (*model.TicketResolution, error) {
	db := r.db

	if tx != nil {
		db = tx.(*gorm.DB)
	}

	if err := db.WithContext(ctx).Create(&resolution).Error; err != nil {
		return nil, err
	}

	return &resolution, nil
}

func (r *TicketResolutionRepo) FindByTicketID(ctx context.Context, ticketID int64) (*model.TicketResolution, error) {
	var resolution model.TicketResolution

	err := r.db.WithContext(ctx).
		Preload("Cause").
		Preload("Solution").
		Where("ticket_id = ?", ticketID).
		First(&resolution).Error

	if err != nil {
		return nil, err
	}

	return &resolution, nil
}
