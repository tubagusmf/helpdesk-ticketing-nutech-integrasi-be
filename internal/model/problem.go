package model

import (
	"context"
	"time"
)

type Problem struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	PartID    int64      `json:"part_id"`
	Part      Part       `json:"part" gorm:"foreignKey:PartID"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type ProblemInput struct {
	Name   string `json:"name"`
	PartID int64  `json:"part_id"`
}

type IProblemRepository interface {
	FindAll(ctx context.Context, search string, page int, limit int) ([]*Problem, int64, error)
	FindByID(ctx context.Context, id int64) (*Problem, error)
	FindByPartID(ctx context.Context, partID int64) ([]*Problem, error)
	ExistsByName(ctx context.Context, name string, partID int64, excludeID int64) (bool, error)
	Create(ctx context.Context, problem *Problem) error
	Update(ctx context.Context, problem *Problem) error
	Delete(ctx context.Context, id int64) error
}

type IProblemUsecase interface {
	FindAll(ctx context.Context, search string, page int, limit int) ([]*Problem, int64, error)
	FindByPartID(ctx context.Context, partID int64) ([]*Problem, error)
	Create(ctx context.Context, in ProblemInput) (*Problem, error)
	Update(ctx context.Context, id int64, in ProblemInput) (*Problem, error)
	Delete(ctx context.Context, id int64) error
}
