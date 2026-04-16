package repository

import (
	"context"
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

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) FindAll(ctx context.Context, filter model.Ticket, search string, startDate string, endDate string, page int, limit int) ([]*model.TicketResponse, int64, error) {
	var tickets []*model.TicketResponse
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Table("tickets").
		Joins("LEFT JOIN projects ON projects.id = tickets.project_id").
		Joins("LEFT JOIN locations ON locations.id = tickets.location_id").
		Joins("LEFT JOIN parts ON parts.id = tickets.part_id").
		Joins("LEFT JOIN asset_ids ON asset_ids.id = tickets.asset_id").
		Joins("LEFT JOIN users as reporter ON reporter.id = tickets.reporter_id").
		Joins("LEFT JOIN users as assigned ON assigned.id = tickets.assigned_to_id").
		Where("tickets.deleted_at IS NULL")

	if search != "" {
		s := "%" + search + "%"

		query = query.Where(`
			(
				tickets.ticket_code ILIKE ?
				OR tickets.description ILIKE ?
				OR reporter.name ILIKE ?
				OR assigned.name ILIKE ?
				OR projects.name ILIKE ?
			)
		`, s, s, s, s, s)
	}

	if filter.ProjectID != 0 {
		query = query.Where("tickets.project_id = ?", filter.ProjectID)
	}

	if filter.AssignedToID != 0 {
		query = query.Where("tickets.assigned_to_id = ?", filter.AssignedToID)
	}

	if filter.ReporterID != 0 {
		query = query.Where("tickets.reporter_id = ?", filter.ReporterID)
	}

	if filter.Priority != "" {
		query = query.Where("tickets.priority = ?", filter.Priority)
	}

	if filter.Status != "" {
		query = query.Where("tickets.status = ?", filter.Status)
	}

	if startDate != "" {
		query = query.Where("DATE(tickets.created_at) >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("DATE(tickets.created_at) <= ?", endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Select(`
			tickets.id,
			tickets.ticket_code,
			tickets.priority,
			tickets.status,
			tickets.description,
			tickets.created_at,
			tickets.due_at,
			tickets.reporter_id,
			tickets.part_id,      
			tickets.asset_id,         
			tickets.attachment,

			projects.name as project_name,
			locations.name as location_name,
			parts.name as part_name,  
			asset_ids.name as asset_code,
			reporter.name as reporter_name,
			assigned.name as assigned_to_name
		`).
		Order("tickets.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&tickets).Error; err != nil {

		return nil, 0, err
	}

	return tickets, total, nil
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
