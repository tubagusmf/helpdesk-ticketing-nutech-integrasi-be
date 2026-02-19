package model

import (
	"context"
	"time"
)

type Cause struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	PartID    int64      `json:"part_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateCauseInput struct {
	Name   string `json:"name" validate:"required,max=100"`
	PartID int64  `json:"part_id" validate:"required"`
}

type UpdateCauseInput struct {
	Name   string `json:"name" validate:"required,max=100"`
	PartID int64  `json:"part_id" validate:"required"`
}

type ICauseRepository interface {
	Create(ctx context.Context, cause Cause) (*Cause, error)
	FindAll(ctx context.Context, filter Cause) ([]*Cause, error)
	FindByID(ctx context.Context, id int64) (*Cause, error)
	Update(ctx context.Context, cause Cause) error
	Delete(ctx context.Context, id int64) error
}

type ICauseUsecase interface {
	Create(ctx context.Context, in CreateCauseInput) (*Cause, error)
	FindAll(ctx context.Context, filter Cause) ([]*Cause, error)
	FindByID(ctx context.Context, id int64) (*Cause, error)
	Update(ctx context.Context, id int64, in UpdateCauseInput) error
	Delete(ctx context.Context, id int64) error
}
