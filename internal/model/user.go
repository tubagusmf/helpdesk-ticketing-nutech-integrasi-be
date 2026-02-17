package model

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ContextAuthKey string

const BearerAuthKey ContextAuthKey = "BearerAuth"

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	RoleID    int64      `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`

	Projects []Project `gorm:"many2many:user_projects;" json:"projects,omitempty"`
}

type IUserRepository interface {
	FindAll(ctx context.Context, user User) ([]*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user User) (*User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id int64) error
}

type IUserUsecase interface {
	FindAll(ctx context.Context, user User) ([]*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Login(ctx context.Context, in LoginInput) (token string, err error)
	Create(ctx context.Context, in CreateUserInput) (token string, err error)
	Update(ctx context.Context, id int64, in UpdateUserInput) error
	Delete(ctx context.Context, id int64) error
}

type LoginInput struct {
	ID       int64  `json:"id"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=3,max=50"`
	RoleID   int64  `json:"role" validate:"required"`
}

type UpdateUserInput struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=3,max=50"`
	RoleID   int64  `json:"role" validate:"required"`
}
