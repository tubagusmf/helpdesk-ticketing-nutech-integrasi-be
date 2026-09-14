package repository

import (
	"context"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type DashboardRepo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) model.IDashboardRepository {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) baseQuery(ctx context.Context, filter model.DashboardFilter) *gorm.DB {
	db := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("tickets.deleted_at IS NULL")

	db = applyFilter(db, filter)

	return db
}

func (r *DashboardRepo) GetSummary(ctx context.Context, filter model.DashboardFilter) (*model.DashboardSummary, error) {
	var result model.DashboardSummary

	base := r.baseQuery(ctx, filter)

	base.Session(&gorm.Session{}).Count(&result.TotalTicket)

	base.Session(&gorm.Session{}).
		Where("status = ? AND due_at < ?", model.StatusOpen, time.Now()).
		Count(&result.SLABreach)

	base.Session(&gorm.Session{}).
		Where("status IN ?", []string{"RESOLVED", "CLOSED"}).
		Count(&result.TicketSelesai)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusOnHold).
		Count(&result.TicketOnHold)

	var avg float64
	base.Session(&gorm.Session{}).
		Where("resolved_at IS NOT NULL").
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600),0)").
		Scan(&avg)

	result.AvgResolutionTime = avg

	return &result, nil
}

func (r *DashboardRepo) GetStatusDistribution(ctx context.Context, filter model.DashboardFilter) (*model.StatusDistribution, error) {
	result := &model.StatusDistribution{}

	base := r.baseQuery(ctx, filter)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusOpen).
		Count(&result.Open)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusInProgress).
		Count(&result.InProgress)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusResolved).
		Count(&result.Resolved)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusClosed).
		Count(&result.Closed)

	base.Session(&gorm.Session{}).
		Where("status = ?", model.StatusOnHold).
		Count(&result.OnHold)

	return result, nil
}

func (r *DashboardRepo) GetPriorityDistribution(ctx context.Context, filter model.DashboardFilter) ([]model.PriorityDistribution, error) {
	var result []model.PriorityDistribution

	err := r.baseQuery(ctx, filter).
		Session(&gorm.Session{}).
		Select("priority, COUNT(*) as total").
		Group("priority").
		Scan(&result).Error

	if err != nil {
		return []model.PriorityDistribution{}, err
	}

	if result == nil {
		return []model.PriorityDistribution{}, nil
	}

	return result, nil
}

func (r *DashboardRepo) GetVolumeProject(ctx context.Context, filter model.DashboardFilter) ([]model.VolumeProject, error) {
	var result []model.VolumeProject

	db := r.db.WithContext(ctx).
		Table("tickets").
		Joins("JOIN projects ON projects.id = tickets.project_id").
		Where("tickets.deleted_at IS NULL")

	db = applyFilter(db, filter)

	err := db.
		Session(&gorm.Session{}).
		Select(`
			projects.name AS project,
			COUNT(tickets.id) AS total
		`).
		Group("projects.id, projects.name").
		Order("projects.name ASC").
		Scan(&result).Error

	if err != nil {
		return []model.VolumeProject{}, err
	}

	if result == nil {
		return []model.VolumeProject{}, nil
	}

	return result, nil
}

func (r *DashboardRepo) GetProjects(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardProject, error) {
	var result []model.DashboardProject

	db := r.db.WithContext(ctx).
		Table("projects").
		Select(`
			projects.id,
			projects.name
		`).
		Where("projects.deleted_at IS NULL")

	switch filter.Role {

	case "EXECUTIVE", "STAFF", "ENGINEER":

		db = db.Where(`
			EXISTS (
				SELECT 1
				FROM user_projects up
				WHERE up.user_id = ?
				AND up.project_id = projects.id
			)
		`, filter.UserID)

	case "USER":

		db = db.Where(`
			EXISTS (
				SELECT 1
				FROM tickets
				WHERE tickets.project_id = projects.id
				AND tickets.reporter_id = ?
				AND tickets.deleted_at IS NULL
			)
		`, filter.UserID)

	case "ADMINISTRATOR":

	default:
		return []model.DashboardProject{}, nil
	}

	err := db.
		Order("projects.name ASC").
		Scan(&result).Error

	if err != nil {
		return []model.DashboardProject{}, err
	}

	if result == nil {
		return []model.DashboardProject{}, nil
	}

	return result, nil
}

func applyFilter(db *gorm.DB, filter model.DashboardFilter) *gorm.DB {
	if filter.ProjectID != 0 {
		db = db.Where("tickets.project_id = ?", filter.ProjectID)
	}

	if filter.PartID != 0 {
		db = db.Where("tickets.part_id = ?", filter.PartID)
	}

	if filter.StartDate != "" {
		db = db.Where("tickets.created_at >= ?", filter.StartDate)
	}

	if filter.EndDate != "" {
		db = db.Where(
			"tickets.created_at < (?::date + INTERVAL '1 day')",
			filter.EndDate,
		)
	}

	// EXECUTIVE
	if filter.Role == "EXECUTIVE" {
		db = db.Where(`
			EXISTS (
				SELECT 1
				FROM user_projects up
				WHERE up.user_id = ?
				AND up.project_id = tickets.project_id
			)
		`, filter.UserID)

		return db
	}

	// STAFF
	if filter.Role == "STAFF" {
		db = db.Where(`
			EXISTS (
				SELECT 1
				FROM user_projects up
				WHERE up.user_id = ?
				AND up.project_id = tickets.project_id
			)
		`, filter.UserID)

		return db
	}

	// USER
	if filter.Role == "USER" {
		db = db.Where(
			"tickets.reporter_id = ?",
			filter.UserID,
		)

		return db
	}

	//ENGINEER
	if filter.Role == "ENGINEER" {
		db = db.Where(`
			EXISTS (
				SELECT 1
				FROM user_projects up
				WHERE up.user_id = ?
				AND up.project_id = tickets.project_id
			)
		`, filter.UserID)

		return db
	}

	// ADMINISTRATOR
	return db
}
