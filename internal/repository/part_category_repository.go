package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type PartCategoryRepo struct {
	db *gorm.DB
}

func NewPartCategoryRepo(db *gorm.DB) model.IPartCategoryRepository {
	return &PartCategoryRepo{db: db}
}

func (r *PartCategoryRepo) Create(ctx context.Context, category model.PartCategory) (*model.PartCategory, error) {
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
		return nil, err
	}

	return r.FindByID(ctx, category.ID)
}

func (r *PartCategoryRepo) FindByID(ctx context.Context, id int64) (*model.PartCategory, error) {
	var category model.PartCategory

	err := r.db.WithContext(ctx).
		Preload("Part").
		Preload("Part.Project").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&category).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("part category not found")
	}

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *PartCategoryRepo) FindAll(ctx context.Context, filter model.PartCategory, page int, limit int) ([]*model.PartCategory, int64, error) {
	var categories []*model.PartCategory
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Model(&model.PartCategory{}).
		Joins("LEFT JOIN parts ON parts.id = part_categories.part_id").
		Where("part_categories.deleted_at IS NULL").
		Preload("Part").
		Preload("Part.Project")

	if filter.Name != "" {
		search := "%" + filter.Name + "%"

		query = query.Where(
			"part_categories.name ILIKE ? OR parts.name ILIKE ?",
			search,
			search,
		)
	}

	if filter.PartID != 0 {
		query = query.Where("part_categories.part_id = ?", filter.PartID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("part_categories.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *PartCategoryRepo) FindByPartID(ctx context.Context, partID int64) ([]*model.PartCategory, error) {
	var categories []*model.PartCategory

	err := r.db.WithContext(ctx).
		Preload("Part").
		Preload("Part.Project").
		Where("part_id = ? AND deleted_at IS NULL", partID).
		Order("name ASC").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *PartCategoryRepo) Update(ctx context.Context, category model.PartCategory) error {
	return r.db.WithContext(ctx).
		Model(&model.PartCategory{}).
		Where("id = ? AND deleted_at IS NULL", category.ID).
		Updates(map[string]interface{}{
			"name":       category.Name,
			"part_id":    category.PartID,
			"updated_at": time.Now(),
		}).Error
}

func (r *PartCategoryRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.PartCategory{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
