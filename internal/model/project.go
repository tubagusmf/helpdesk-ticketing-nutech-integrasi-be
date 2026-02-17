package model

import (
	"context"
	"time"
)

type Project struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`

	Users []User `gorm:"many2many:user_projects;" json:"users,omitempty"`
}

type CreateProjectInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateProjectInput struct {
	Name string `json:"name" validate:"required"`
}

type IProjectRepository interface {
	FindAll(ctx context.Context, project Project) ([]*Project, error)
	FindByID(ctx context.Context, id int64) (*Project, error)
	Create(ctx context.Context, project Project) (*Project, error)
	Update(ctx context.Context, project Project) error
	Delete(ctx context.Context, id int64) error
	AssignUser(ctx context.Context, userID, projectID int64) error
	RemoveUser(ctx context.Context, userID, projectID int64) error
}

type IProjectUsecase interface {
	FindAll(ctx context.Context, project Project) ([]*Project, error)
	FindByID(ctx context.Context, id int64) (*Project, error)
	Create(ctx context.Context, in CreateProjectInput) (*Project, error)
	Update(ctx context.Context, id int64, in UpdateProjectInput) error
	Delete(ctx context.Context, id int64) error
}
