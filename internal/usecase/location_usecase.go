package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type LocationUsecase struct {
	locationRepo model.ILocationRepository
}

func NewLocationUsecase(locationRepo model.ILocationRepository) model.ILocationUsecase {
	return &LocationUsecase{
		locationRepo: locationRepo,
	}
}

func (u *LocationUsecase) Create(ctx context.Context, in model.LocationInput) (*model.Location, error) {
	log := logrus.WithFields(logrus.Fields{
		"in": in,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	location := model.Location{
		Name:      in.Name,
		ProjectID: in.ProjectID,
	}

	created, err := u.locationRepo.Create(ctx, location)
	if err != nil {
		log.Error("Failed to create location: ", err)
		return nil, err
	}

	return created, nil
}

func (u *LocationUsecase) FindAll(ctx context.Context, filter model.Location) ([]*model.Location, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	locations, err := u.locationRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch locations: ", err)
		return nil, err
	}

	return locations, nil
}

func (u *LocationUsecase) FindByID(ctx context.Context, id int64) (*model.Location, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	location, err := u.locationRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find location: ", err)
		return nil, err
	}

	return location, nil
}

func (u *LocationUsecase) Update(ctx context.Context, id int64, in model.UpdateLocationInput) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	location, err := u.locationRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Location not found: ", err)
		return err
	}

	if location == nil {
		return errors.New("location not found")
	}

	location.Name = in.Name
	location.ProjectID = in.ProjectID

	if err := u.locationRepo.Update(ctx, *location); err != nil {
		log.Error("Failed to update location: ", err)
		return err
	}

	return nil
}

func (u *LocationUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	location, err := u.locationRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find location for deletion: ", err)
		return err
	}

	if location == nil {
		return errors.New("location not found")
	}

	if err := u.locationRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete location: ", err)
		return err
	}

	return nil
}
