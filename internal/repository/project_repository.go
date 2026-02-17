package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) model.IProjectRepository {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, project model.Project) (*model.Project, error) {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	if err := r.db.WithContext(ctx).Create(&project).Error; err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepo) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&project).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("project not found")
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepo) FindAll(ctx context.Context, filter model.Project) ([]*model.Project, error) {
	var projects []*model.Project

	query := r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("deleted_at IS NULL")

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if err := query.Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepo) Update(ctx context.Context, project model.Project) error {
	project.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("id = ? AND deleted_at IS NULL", project.ID).
		Updates(project).Error
}

func (r *ProjectRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Project{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

func (r *ProjectRepo) AssignUser(ctx context.Context, userID, projectID int64) error {
	var project model.Project

	if err := r.db.WithContext(ctx).
		First(&project, "id = ?", projectID).Error; err != nil {
		return err
	}

	var user model.User
	if err := r.db.WithContext(ctx).
		First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Model(&project).
		Association("Users").
		Append(&user)
}

func (r *ProjectRepo) RemoveUser(ctx context.Context, userID, projectID int64) error {
	var project model.Project
	if err := r.db.WithContext(ctx).
		First(&project, "id = ?", projectID).Error; err != nil {
		return err
	}

	var user model.User
	if err := r.db.WithContext(ctx).
		First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Model(&project).
		Association("Users").
		Delete(&user)
}
