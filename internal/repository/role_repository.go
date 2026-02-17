package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type RoleRepo struct {
	db *gorm.DB
}

func NewRoleRepo(db *gorm.DB) model.IRoleRepository {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) Create(ctx context.Context, role model.Role) (*model.Role, error) {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&role).Error; err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepo) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	var role model.Role

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&role).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role not found")
	}

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepo) FindAll(ctx context.Context, filter model.Role) ([]*model.Role, error) {
	var roles []*model.Role

	query := r.db.WithContext(ctx).
		Model(&model.Role{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Privilege != "" {
		query = query.Where("privilege ILIKE ?", "%"+filter.Privilege+"%")
	}

	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RoleRepo) Update(ctx context.Context, role model.Role) error {
	role.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Role{}).
		Where("id = ? AND deleted_at IS NULL", role.ID).
		Updates(role).Error
}

func (r *RoleRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Role{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}
