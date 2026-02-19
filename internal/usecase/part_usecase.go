package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type PartUsecase struct {
	partRepo model.IPartRepository
}

func NewPartUsecase(partRepo model.IPartRepository) model.IPartUsecase {
	return &PartUsecase{partRepo: partRepo}
}

func (u *PartUsecase) Create(ctx context.Context, in model.PartInput) (*model.Part, error) {
	log := logrus.WithFields(logrus.Fields{"in": in})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	part := model.Part{
		Name:      in.Name,
		ProjectID: in.ProjectID,
	}

	created, err := u.partRepo.Create(ctx, part)
	if err != nil {
		log.Error("Failed to create part: ", err)
		return nil, err
	}

	return created, nil
}

func (u *PartUsecase) FindAll(ctx context.Context, filter model.Part) ([]*model.Part, error) {
	log := logrus.WithFields(logrus.Fields{"filter": filter})

	locations, err := u.partRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch parts: ", err)
		return nil, err
	}

	return locations, nil
}

func (u *PartUsecase) FindByID(ctx context.Context, id int64) (*model.Part, error) {
	log := logrus.WithFields(logrus.Fields{"id": id})

	part, err := u.partRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find part: ", err)
		return nil, err
	}

	return part, nil
}

func (u *PartUsecase) Update(ctx context.Context, id int64, in model.UpdatePartInput) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	part, err := u.partRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if part == nil {
		return errors.New("part not found")
	}

	part.Name = in.Name
	part.ProjectID = in.ProjectID

	if err := u.partRepo.Update(ctx, *part); err != nil {
		log.Error("Failed to update part: ", err)
		return err
	}

	return nil
}

func (u *PartUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	part, err := u.partRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find part for deletion: ", err)
		return err
	}

	if part == nil {
		return errors.New("part not found")
	}

	if err := u.partRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete part: ", err)
		return err
	}

	return nil
}
