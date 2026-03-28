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

func (r *TicketRepo) FindAll(ctx context.Context, filter model.Ticket) ([]*model.TicketResponse, error) {
	var tickets []*model.TicketResponse

	err := r.db.WithContext(ctx).
		Table("tickets").
		Select(`
			tickets.id,
			tickets.ticket_code,
			tickets.priority,
			tickets.status,
			tickets.description,
			tickets.created_at,
			tickets.due_at,

			projects.name as project_name,
			locations.name as location_name,
			asset_ids.name as asset_code,
			users.name as reporter_name
		`).
		Joins("LEFT JOIN projects ON projects.id = tickets.project_id").
		Joins("LEFT JOIN locations ON locations.id = tickets.location_id").
		Joins("LEFT JOIN asset_ids ON asset_ids.id = tickets.asset_id").
		Joins("LEFT JOIN users ON users.id = tickets.reporter_id").
		Where("tickets.deleted_at IS NULL").
		Order("tickets.created_at DESC").
		Scan(&tickets).Error

	return tickets, err
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
