package model

import (
	"context"
	"time"
)

type Part struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	ProjectID int64      `json:"project_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type PartInput struct {
	Name      string `json:"name" validate:"required"`
	ProjectID int64  `json:"project_id" validate:"required"`
}

type UpdatePartInput struct {
	Name      string `json:"name" validate:"required"`
	ProjectID int64  `json:"project_id" validate:"required"`
}

type IPartRepository interface {
	FindAll(ctx context.Context, filter Part) ([]*Part, error)
	FindByID(ctx context.Context, id int64) (*Part, error)
	Create(ctx context.Context, part Part) (*Part, error)
	Update(ctx context.Context, part Part) error
	Delete(ctx context.Context, id int64) error
}

type IPartUsecase interface {
	FindAll(ctx context.Context, filter Part) ([]*Part, error)
	FindByID(ctx context.Context, id int64) (*Part, error)
	Create(ctx context.Context, in PartInput) (*Part, error)
	Update(ctx context.Context, id int64, in UpdatePartInput) error
	Delete(ctx context.Context, id int64) error
}
