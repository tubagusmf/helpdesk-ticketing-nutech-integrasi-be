package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) model.IUserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user model.User) (*model.User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) FindAll(ctx context.Context, filter model.User) ([]*model.User, error) {
	var users []*model.User

	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("deleted_at IS NULL")

	if filter.Email != "" {
		query = query.Where("email ILIKE ?", "%"+filter.Email+"%")
	}

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepo) Update(ctx context.Context, user model.User) error {
	user.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(user).Error
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}
