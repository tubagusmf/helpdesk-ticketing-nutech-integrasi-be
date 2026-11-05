package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type PartCategoryUsecase struct {
	categoryRepo model.IPartCategoryRepository
	partRepo     model.IPartRepository
}

func NewPartCategoryUsecase(categoryRepo model.IPartCategoryRepository, partRepo model.IPartRepository) model.IPartCategoryUsecase {
	return &PartCategoryUsecase{
		categoryRepo: categoryRepo,
		partRepo:     partRepo,
	}
}

func (u *PartCategoryUsecase) Create(ctx context.Context, in model.PartCategoryInput) (*model.PartCategory, error) {
	if err := validate.Struct(in); err != nil {
		return nil, err
	}

	if _, err := u.partRepo.FindByID(ctx, in.PartID); err != nil {
		return nil, errors.New("part not found")
	}

	category := model.PartCategory{
		Name:   in.Name,
		PartID: in.PartID,
	}

	created, err := u.categoryRepo.Create(ctx, category)
	if err != nil {
		logrus.Error("Failed to create part category: ", err)
		return nil, err
	}

	return created, nil
}

func (u *PartCategoryUsecase) FindAll(ctx context.Context, filter model.PartCategory, page int, limit int) ([]*model.PartCategory, int64, error) {
	categories, total, err := u.categoryRepo.FindAll(
		ctx,
		filter,
		page,
		limit,
	)

	if err != nil {
		logrus.Error("Failed to fetch part categories: ", err)
		return nil, 0, err
	}

	return categories, total, nil
}

func (u *PartCategoryUsecase) FindByID(ctx context.Context, id int64) (*model.PartCategory, error) {
	category, err := u.categoryRepo.FindByID(ctx, id)
	if err != nil {
		logrus.Error("Failed to find part category: ", err)
		return nil, err
	}

	return category, nil
}

func (u *PartCategoryUsecase) FindByPartID(ctx context.Context, partID int64) ([]*model.PartCategory, error) {
	categories, err := u.categoryRepo.FindByPartID(ctx, partID)
	if err != nil {
		logrus.Error("Failed to fetch categories by part ID: ", err)
		return nil, err
	}

	return categories, nil
}

func (u *PartCategoryUsecase) Update(ctx context.Context, id int64, in model.UpdatePartCategoryInput) error {
	if err := validate.Struct(in); err != nil {
		return err
	}

	if _, err := u.categoryRepo.FindByID(ctx, id); err != nil {
		return err
	}

	if _, err := u.partRepo.FindByID(ctx, in.PartID); err != nil {
		return errors.New("part not found")
	}

	category := model.PartCategory{
		ID:     id,
		Name:   in.Name,
		PartID: in.PartID,
	}

	if err := u.categoryRepo.Update(ctx, category); err != nil {
		logrus.Error("Failed to update part category: ", err)
		return err
	}

	return nil
}

func (u *PartCategoryUsecase) Delete(ctx context.Context, id int64) error {
	if _, err := u.categoryRepo.FindByID(ctx, id); err != nil {
		return err
	}

	if err := u.categoryRepo.Delete(ctx, id); err != nil {
		logrus.Error("Failed to delete part category: ", err)
		return err
	}

	return nil
}
