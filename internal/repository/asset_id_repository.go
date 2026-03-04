package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type AssetIDRepo struct {
	db *gorm.DB
}

func NewAssetIDRepo(db *gorm.DB) model.IAssetIDRepository {
	return &AssetIDRepo{db: db}
}

func (r *AssetIDRepo) Create(ctx context.Context, asset model.AssetID) (*model.AssetID, error) {
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&asset).Error; err != nil {
		return nil, err
	}

	return &asset, nil
}

func (r *AssetIDRepo) FindByID(ctx context.Context, id int64) (*model.AssetID, error) {
	var asset model.AssetID

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&asset).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("asset_id not found")
	}

	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (r *AssetIDRepo) FindAll(ctx context.Context, filter model.AssetID, page int, limit int) ([]*model.AssetID, int64, error) {
	var assets []*model.AssetID
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.AssetID{}).
		Joins("LEFT JOIN parts ON parts.id = asset_ids.part_id").
		Joins("LEFT JOIN projects ON projects.id = parts.project_id").
		Where("asset_ids.deleted_at IS NULL").
		Preload("Part").
		Preload("Part.Project")

	if filter.Name != "" {
		search := "%" + filter.Name + "%"
		query = query.Where(`
			asset_ids.name ILIKE ?
			OR parts.name ILIKE ?
			OR projects.name ILIKE ?
		`, search, search, search)
	}

	if filter.PartID != 0 {
		query = query.Where("asset_ids.part_id = ?", filter.PartID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(limit).
		Offset((page - 1) * limit).
		Order("asset_ids.id DESC").
		Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *AssetIDRepo) Update(ctx context.Context, asset model.AssetID) error {
	asset.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.AssetID{}).
		Where("id = ? AND deleted_at IS NULL", asset.ID).
		Updates(map[string]interface{}{
			"name":       asset.Name,
			"part_id":    asset.PartID,
			"updated_at": asset.UpdatedAt,
		}).Error
}

func (r *AssetIDRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.AssetID{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
