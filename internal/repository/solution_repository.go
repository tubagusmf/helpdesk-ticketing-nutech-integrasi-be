package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type SolutionRepo struct {
	db *gorm.DB
}

func NewSolutionRepo(db *gorm.DB) model.ISolutionRepository {
	return &SolutionRepo{db: db}
}

func (r *SolutionRepo) Create(ctx context.Context, solution model.Solution) (*model.Solution, error) {
	solution.CreatedAt = time.Now()
	solution.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&solution).Error; err != nil {
		return nil, err
	}

	return &solution, nil
}

func (r *SolutionRepo) FindByID(ctx context.Context, id int64) (*model.Solution, error) {
	var solution model.Solution

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&solution).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("solution not found")
	}

	if err != nil {
		return nil, err
	}

	return &solution, nil
}

func (r *SolutionRepo) FindAll(ctx context.Context, filter model.Solution) ([]*model.Solution, error) {
	var solutions []*model.Solution

	query := r.db.WithContext(ctx).
		Model(&model.Solution{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.CauseID != 0 {
		query = query.Where("cause_id = ?", filter.CauseID)
	}

	if err := query.Find(&solutions).Error; err != nil {
		return nil, err
	}

	return solutions, nil
}

func (r *SolutionRepo) Update(ctx context.Context, solution model.Solution) error {
	solution.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Solution{}).
		Where("id = ? AND deleted_at IS NULL", solution.ID).
		Updates(map[string]interface{}{
			"name":       solution.Name,
			"cause_id":   solution.CauseID,
			"updated_at": solution.UpdatedAt,
		}).Error
}

func (r *SolutionRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Solution{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}
