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

func (r *CauseRepo) FindAll(ctx context.Context, filter model.Cause) ([]*model.Cause, error) {
	var causes []*model.Cause

	query := r.db.WithContext(ctx).
		Model(&model.Cause{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.PartID != 0 {
		query = query.Where("part_id = ?", filter.PartID)
	}

	if err := query.Find(&causes).Error; err != nil {
		return nil, err
	}

	return causes, nil
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
