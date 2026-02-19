package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type LocationRepo struct {
	db *gorm.DB
}

func NewLocationRepo(db *gorm.DB) model.ILocationRepository {
	return &LocationRepo{db: db}
}

func (r *LocationRepo) Create(ctx context.Context, location model.Location) (*model.Location, error) {
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&location).Error; err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *LocationRepo) FindByID(ctx context.Context, id int64) (*model.Location, error) {
	var location model.Location

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&location).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("location not found")
	}

	if err != nil {
		return nil, err
	}

	return &location, nil
}

func (r *LocationRepo) FindAll(ctx context.Context, filter model.Location) ([]*model.Location, error) {
	var locations []*model.Location

	query := r.db.WithContext(ctx).
		Model(&model.Location{}).
		Where("deleted_at IS NULL")

	// Optional filter by name
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	// Optional filter by project_id
	if filter.ProjectID != 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if err := query.Find(&locations).Error; err != nil {
		return nil, err
	}

	return locations, nil
}

func (r *LocationRepo) Update(ctx context.Context, location model.Location) error {
	location.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Location{}).
		Where("id = ? AND deleted_at IS NULL", location.ID).
		Updates(map[string]interface{}{
			"name":       location.Name,
			"project_id": location.ProjectID,
			"updated_at": location.UpdatedAt,
		}).Error
}

func (r *LocationRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Location{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
