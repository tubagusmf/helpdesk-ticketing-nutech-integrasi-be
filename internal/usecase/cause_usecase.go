package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type CauseUsecase struct {
	causeRepo model.ICauseRepository
}

func NewCauseUsecase(causeRepo model.ICauseRepository) model.ICauseUsecase {
	return &CauseUsecase{
		causeRepo: causeRepo,
	}
}

func (u *CauseUsecase) Create(ctx context.Context, in model.CreateCauseInput) (*model.Cause, error) {
	log := logrus.WithField("input", in)

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	cause := model.Cause{
		Name:   in.Name,
		PartID: in.PartID,
	}

	created, err := u.causeRepo.Create(ctx, cause)
	if err != nil {
		log.Error("Failed to create cause: ", err)
		return nil, err
	}

	return created, nil
}

func (u *CauseUsecase) FindAll(ctx context.Context, filter model.Cause) ([]*model.Cause, error) {
	log := logrus.WithFields(logrus.Fields{"filter": filter})

	causes, err := u.causeRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch causes: ", err)
		return nil, err
	}

	return causes, nil
}

func (u *CauseUsecase) FindByID(ctx context.Context, id int64) (*model.Cause, error) {
	log := logrus.WithFields(logrus.Fields{"id": id})

	cause, err := u.causeRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find cause: ", err)
		return nil, err
	}

	return cause, nil
}

func (u *CauseUsecase) Update(ctx context.Context, id int64, in model.UpdateCauseInput) error {
	log := logrus.WithField("id", id)

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	cause, err := u.causeRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if cause == nil {
		return errors.New("cause not found")
	}

	cause.Name = in.Name
	cause.PartID = in.PartID

	if err := u.causeRepo.Update(ctx, *cause); err != nil {
		log.Error("Failed to update cause: ", err)
		return err
	}

	return nil
}

func (u *CauseUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	cause, err := u.causeRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find cause for deletion: ", err)
		return err
	}

	if cause == nil {
		log.Error("Failed to find cause for deletion: ", err)
		return errors.New("cause not found")
	}

	if err := u.causeRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete cause: ", err)
		return err
	}

	return nil
}
