package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type PartRepo struct {
	db *gorm.DB
}

func NewPartRepo(db *gorm.DB) model.IPartRepository {
	return &PartRepo{db: db}
}

func (r *PartRepo) Create(ctx context.Context, part model.Part) (*model.Part, error) {
	part.CreatedAt = time.Now()
	part.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&part).Error; err != nil {
		return nil, err
	}

	return &part, nil
}

func (r *PartRepo) FindByID(ctx context.Context, id int64) (*model.Part, error) {
	var part model.Part

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&part).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("part not found")
	}

	if err != nil {
		return nil, err
	}

	return &part, nil
}

func (r *PartRepo) FindAll(ctx context.Context, filter model.Part) ([]*model.Part, error) {
	var parts []*model.Part

	query := r.db.WithContext(ctx).
		Model(&model.Part{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.ProjectID != 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if err := query.Find(&parts).Error; err != nil {
		return nil, err
	}

	return parts, nil
}

func (r *PartRepo) Update(ctx context.Context, part model.Part) error {
	part.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Part{}).
		Where("id = ? AND deleted_at IS NULL", part.ID).
		Updates(map[string]interface{}{
			"name":       part.Name,
			"project_id": part.ProjectID,
			"updated_at": part.UpdatedAt,
		}).Error
}

func (r *PartRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Part{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
