package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type CauseRepo struct {
	db *gorm.DB
}

func NewCauseRepo(db *gorm.DB) model.ICauseRepository {
	return &CauseRepo{db: db}
}

func (r *CauseRepo) Create(ctx context.Context, cause model.Cause) (*model.Cause, error) {
	cause.CreatedAt = time.Now()
	cause.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&cause).Error; err != nil {
		return nil, err
	}

	return &cause, nil
}

func (r *CauseRepo) FindByID(ctx context.Context, id int64) (*model.Cause, error) {
	var cause model.Cause

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&cause).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("cause not found")
	}

	if err != nil {
		return nil, err
	}

	return &cause, nil
}

func (r *CauseRepo) FindAll(ctx context.Context, filter model.Cause, page int, limit int) ([]*model.Cause, int64, error) {
	var causes []*model.Cause
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Cause{}).
		Joins("LEFT JOIN parts ON parts.id = causes.part_id").
		Joins("LEFT JOIN projects ON projects.id = parts.project_id").
		Where("causes.deleted_at IS NULL").
		Preload("Part").
		Preload("Part.Project")

	if filter.Name != "" {
		search := "%" + filter.Name + "%"
		query = query.Where(`causes.name ILIKE ? OR parts.name ILIKE ? OR projects.name ILIKE ?`, search, search, search)
	}

	if filter.PartID != 0 {
		query = query.Where("causes.part_id = ?", filter.PartID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&causes).Error; err != nil {
		return nil, 0, err
	}

	return causes, total, nil
}

func (r *CauseRepo) Update(ctx context.Context, cause model.Cause) error {
	cause.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Cause{}).
		Where("id = ? AND deleted_at IS NULL", cause.ID).
		Updates(map[string]interface{}{
			"name":       cause.Name,
			"part_id":    cause.PartID,
			"updated_at": cause.UpdatedAt,
		}).Error
}

func (r *CauseRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Cause{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
