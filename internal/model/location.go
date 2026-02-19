package model

import (
	"context"
	"time"
)

type Location struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	ProjectID int64      `json:"project_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type LocationInput struct {
	Name      string `json:"name" validate:"required"`
	ProjectID int64  `json:"project_id" validate:"required"`
}

type UpdateLocationInput struct {
	Name      string `json:"name" validate:"required"`
	ProjectID int64  `json:"project_id" validate:"required"`
}

type ILocationRepository interface {
	FindAll(ctx context.Context, location Location) ([]*Location, error)
	FindByID(ctx context.Context, id int64) (*Location, error)
	Create(ctx context.Context, location Location) (*Location, error)
	Update(ctx context.Context, location Location) error
	Delete(ctx context.Context, id int64) error
}

type ILocationUsecase interface {
	FindAll(ctx context.Context, location Location) ([]*Location, error)
	FindByID(ctx context.Context, id int64) (*Location, error)
	Create(ctx context.Context, in LocationInput) (*Location, error)
	Update(ctx context.Context, id int64, in UpdateLocationInput) error
	Delete(ctx context.Context, id int64) error
}
