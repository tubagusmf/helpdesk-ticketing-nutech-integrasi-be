package model

import (
	"context"
	"time"
)

type PartCategory struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	PartID    int64      `json:"part_id"`
	Part      Part       `json:"part" gorm:"foreignKey:PartID"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type PartCategoryInput struct {
	Name   string `json:"name" validate:"required"`
	PartID int64  `json:"part_id" validate:"required"`
}

type UpdatePartCategoryInput struct {
	Name   string `json:"name" validate:"required"`
	PartID int64  `json:"part_id" validate:"required"`
}

type IPartCategoryRepository interface {
	FindAll(ctx context.Context, filter PartCategory, page int, limit int) ([]*PartCategory, int64, error)
	FindByID(ctx context.Context, id int64) (*PartCategory, error)
	FindByPartID(ctx context.Context, partID int64) ([]*PartCategory, error)
	Create(ctx context.Context, category PartCategory) (*PartCategory, error)
	Update(ctx context.Context, category PartCategory) error
	Delete(ctx context.Context, id int64) error
}

type IPartCategoryUsecase interface {
	FindAll(ctx context.Context, filter PartCategory, page int, limit int) ([]*PartCategory, int64, error)
	FindByID(ctx context.Context, id int64) (*PartCategory, error)
	FindByPartID(ctx context.Context, partID int64) ([]*PartCategory, error)
	Create(ctx context.Context, in PartCategoryInput) (*PartCategory, error)
	Update(ctx context.Context, id int64, in UpdatePartCategoryInput) error
	Delete(ctx context.Context, id int64) error
}
