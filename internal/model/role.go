package model

import (
	"context"
	"time"
)

type Role struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Privilege string     `json:"privilege"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type CreateRoleInput struct {
	Name      string `json:"name" validate:"required"`
	Privilege string `json:"privilege" validate:"required"`
}

type UpdateRoleInput struct {
	Name      string `json:"name" validate:"required"`
	Privilege string `json:"privilege" validate:"required"`
}

type IRoleRepository interface {
	FindAll(ctx context.Context, role Role) ([]*Role, error)
	FindByID(ctx context.Context, id int64) (*Role, error)
	Create(ctx context.Context, role Role) (*Role, error)
	Update(ctx context.Context, role Role) error
	Delete(ctx context.Context, id int64) error
}

type IRoleUsecase interface {
	FindAll(ctx context.Context, role Role) ([]*Role, error)
	FindByID(ctx context.Context, id int64) (*Role, error)
	Create(ctx context.Context, in CreateRoleInput) (*Role, error)
	Update(ctx context.Context, id int64, in UpdateRoleInput) error
	Delete(ctx context.Context, id int64) error
}
