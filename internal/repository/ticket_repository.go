package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketRepo struct {
	db *gorm.DB
}

func NewTicketRepo(db *gorm.DB) model.ITicketRepository {
	return &TicketRepo{db: db}
}

func (r *TicketRepo) Create(ctx context.Context, ticket model.Ticket) (*model.Ticket, error) {
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	ticket.Status = model.StatusOpen

	if err := r.db.WithContext(ctx).Create(&ticket).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) FindByID(ctx context.Context, id int64) (*model.Ticket, error) {
	var ticket model.Ticket

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&ticket).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("ticket not found")
	}

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) FindAll(ctx context.Context, filter model.Ticket) ([]*model.Ticket, error) {
	var tickets []*model.Ticket

	query := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("deleted_at IS NULL")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}

	if err := query.Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *TicketRepo) Update(ctx context.Context, ticket model.Ticket) error {
	ticket.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ? AND deleted_at IS NULL", ticket.ID).
		Updates(ticket).Error
}

func (r *TicketRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}
