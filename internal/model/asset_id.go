package model

import (
	"context"
	"time"
)

type AssetID struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	PartID    int64      `json:"part_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type AssetIDInput struct {
	Name   string `json:"name" validate:"required"`
	PartID int64  `json:"part_id" validate:"required"`
}

type UpdateAssetIDInput struct {
	Name   string `json:"name" validate:"required"`
	PartID int64  `json:"part_id" validate:"required"`
}

type IAssetIDRepository interface {
	FindAll(ctx context.Context, filter AssetID) ([]*AssetID, error)
	FindByID(ctx context.Context, id int64) (*AssetID, error)
	Create(ctx context.Context, asset AssetID) (*AssetID, error)
	Update(ctx context.Context, asset AssetID) error
	Delete(ctx context.Context, id int64) error
}

type IAssetIDUsecase interface {
	FindAll(ctx context.Context, filter AssetID) ([]*AssetID, error)
	FindByID(ctx context.Context, id int64) (*AssetID, error)
	Create(ctx context.Context, in AssetIDInput) (*AssetID, error)
	Update(ctx context.Context, id int64, in UpdateAssetIDInput) error
	Delete(ctx context.Context, id int64) error
}
