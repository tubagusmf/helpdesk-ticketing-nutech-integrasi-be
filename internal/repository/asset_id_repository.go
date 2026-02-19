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

func (r *AssetIDRepo) FindAll(ctx context.Context, filter model.AssetID) ([]*model.AssetID, error) {
	var assets []*model.AssetID

	query := r.db.WithContext(ctx).
		Model(&model.AssetID{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.PartID != 0 {
		query = query.Where("part_id = ?", filter.PartID)
	}

	if err := query.Find(&assets).Error; err != nil {
		return nil, err
	}

	return assets, nil
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
