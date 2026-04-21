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

func (r *DashboardRepo) baseQuery(ctx context.Context, filter map[string]interface{}) *gorm.DB {
	return applyFilter(
		r.db.WithContext(ctx).
			Model(&model.Ticket{}).
			Where("deleted_at IS NULL"),
		filter,
	)
}

func (r *DashboardRepo) GetSummary(ctx context.Context, filter map[string]interface{}) (*model.DashboardSummary, error) {
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

func (r *DashboardRepo) GetStatusDistribution(ctx context.Context, filter map[string]interface{}) (*model.StatusDistribution, error) {
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

func (r *DashboardRepo) GetPriorityDistribution(ctx context.Context, filter map[string]interface{}) ([]model.PriorityDistribution, error) {
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

func (r *DashboardRepo) GetVolumeProject(ctx context.Context, filter map[string]interface{}) ([]model.VolumeProject, error) {
	var result []model.VolumeProject

	db := r.db.WithContext(ctx).
		Table("tickets").
		Joins("JOIN projects ON projects.id = tickets.project_id").
		Where("tickets.deleted_at IS NULL")

	db = applyFilter(db, filter)

	err := db.
		Session(&gorm.Session{}).
		Select("projects.name as project, COUNT(tickets.id) as total").
		Group("projects.name").
		Scan(&result).Error

	if err != nil {
		return []model.VolumeProject{}, err
	}

	if result == nil {
		return []model.VolumeProject{}, nil
	}

	return result, nil
}

func applyFilter(db *gorm.DB, filter map[string]interface{}) *gorm.DB {
	if v, ok := filter["project_id"]; ok && v != "" {
		db = db.Where("project_id = ?", v)
	}

	if v, ok := filter["part_id"]; ok && v != "" {
		db = db.Where("part_id = ?", v)
	}

	if v, ok := filter["start_date"]; ok && v != "" {
		db = db.Where("created_at >= ?", v)
	}

	if v, ok := filter["end_date"]; ok && v != "" {
		db = db.Where("created_at <= ?", v)
	}

	return db
}
