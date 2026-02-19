package model

import (
	"context"
	"time"
)

type Solution struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name"`
	CauseID   int64      `json:"cause_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type CreateSolutionInput struct {
	Name    string `json:"name" validate:"required,max=100"`
	CauseID int64  `json:"cause_id" validate:"required"`
}

type UpdateSolutionInput struct {
	Name    string `json:"name" validate:"required,max=100"`
	CauseID int64  `json:"cause_id" validate:"required"`
}

type ISolutionRepository interface {
	Create(ctx context.Context, solution Solution) (*Solution, error)
	FindByID(ctx context.Context, id int64) (*Solution, error)
	FindAll(ctx context.Context, filter Solution) ([]*Solution, error)
	Update(ctx context.Context, solution Solution) error
	Delete(ctx context.Context, id int64) error
}

type ISolutionUsecase interface {
	Create(ctx context.Context, in CreateSolutionInput) (*Solution, error)
	FindByID(ctx context.Context, id int64) (*Solution, error)
	FindAll(ctx context.Context, filter Solution) ([]*Solution, error)
	Update(ctx context.Context, id int64, in UpdateSolutionInput) error
	Delete(ctx context.Context, id int64) error
}
