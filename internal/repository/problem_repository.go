package repository

import (
	"context"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type ProblemRepo struct {
	db *gorm.DB
}

func NewProblemRepo(db *gorm.DB) model.IProblemRepository {
	return &ProblemRepo{db: db}
}

func (r *ProblemRepo) FindAll(ctx context.Context, search string, page int, limit int) ([]*model.Problem, int64, error) {
	var data []*model.Problem
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Problem{}).
		Where("deleted_at IS NULL")

	if search != "" {
		query = query.Where(
			"name ILIKE ?",
			"%"+search+"%",
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if err := query.
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&data).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (r *ProblemRepo) FindByID(ctx context.Context, id int64) (*model.Problem, error) {
	var problem model.Problem

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&problem).Error

	if err != nil {
		return nil, err
	}

	return &problem, nil
}

func (r *ProblemRepo) ExistsByName(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Problem{}).
		Where("LOWER(name) = LOWER(?)", name).
		Where("deleted_at IS NULL")

	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ProblemRepo) Create(ctx context.Context, problem *model.Problem) error {
	now := time.Now()

	problem.CreatedAt = now
	problem.UpdatedAt = now

	return r.db.WithContext(ctx).
		Create(problem).Error
}

func (r *ProblemRepo) Update(ctx context.Context, problem *model.Problem) error {
	result := r.db.WithContext(ctx).
		Model(&model.Problem{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			problem.ID,
		).
		Updates(map[string]interface{}{
			"name":       problem.Name,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ProblemRepo) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Problem{}).
		Where(
			"id = ? AND deleted_at IS NULL",
			id,
		).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
