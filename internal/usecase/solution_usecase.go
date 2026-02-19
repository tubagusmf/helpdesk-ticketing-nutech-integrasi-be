package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type SolutionUsecase struct {
	solutionRepo model.ISolutionRepository
}

func NewSolutionUsecase(solutionRepo model.ISolutionRepository) model.ISolutionUsecase {
	return &SolutionUsecase{solutionRepo: solutionRepo}
}

func (u *SolutionUsecase) Create(ctx context.Context, in model.CreateSolutionInput) (*model.Solution, error) {
	log := logrus.WithField("input", in)

	if err := validate.Struct(in); err != nil {
		log.Error("validation error:", err)
		return nil, err
	}

	solution := model.Solution{
		Name:    in.Name,
		CauseID: in.CauseID,
	}

	created, err := u.solutionRepo.Create(ctx, solution)
	if err != nil {
		log.Error("failed to create solution: ", err)
		return nil, err
	}

	return created, nil
}

func (u *SolutionUsecase) FindByID(ctx context.Context, id int64) (*model.Solution, error) {
	log := logrus.WithFields(logrus.Fields{"id": id})

	solution, err := u.solutionRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find solution: ", err)
		return nil, err
	}

	return solution, nil
}

func (u *SolutionUsecase) FindAll(ctx context.Context, filter model.Solution) ([]*model.Solution, error) {
	log := logrus.WithFields(logrus.Fields{"filter": filter})

	solutions, err := u.solutionRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch solutions: ", err)
		return nil, err
	}

	return solutions, nil
}

func (u *SolutionUsecase) Update(ctx context.Context, id int64, in model.UpdateSolutionInput) error {
	log := logrus.WithFields(logrus.Fields{"id": id, "input": in})

	if err := validate.Struct(in); err != nil {
		log.Error("validation error:", err)
		return err
	}

	solution, err := u.solutionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if solution == nil {
		return errors.New("solution not found")
	}

	solution.Name = in.Name
	solution.CauseID = in.CauseID

	if err := u.solutionRepo.Update(ctx, *solution); err != nil {
		log.Error("Failed to update solution: ", err)
		return err
	}

	return nil
}

func (u *SolutionUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{"id": id})

	solution, err := u.solutionRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find solution for deletion: ", err)
		return err
	}

	if solution == nil {
		return errors.New("solution not found")
	}

	if err := u.solutionRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete solution: ", err)
		return err
	}

	return nil
}
