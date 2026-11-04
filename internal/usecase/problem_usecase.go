package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type ProblemUsecase struct {
	problemRepo model.IProblemRepository
}

func NewProblemUsecase(problemRepo model.IProblemRepository) model.IProblemUsecase {
	return &ProblemUsecase{
		problemRepo: problemRepo,
	}
}

func validateProblemName(name string) (string, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", errors.New(
			"nama permasalahan wajib diisi",
		)
	}

	if !utf8.ValidString(name) {
		return "", errors.New(
			"nama permasalahan tidak valid",
		)
	}

	if utf8.RuneCountInString(name) > 255 {
		return "", errors.New(
			"nama permasalahan maksimal 255 karakter",
		)
	}

	return name, nil
}

func (u *ProblemUsecase) FindAll(ctx context.Context, search string, page int, limit int) ([]*model.Problem, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit != 10 && limit != 25 && limit != 50 {
		limit = 10
	}

	return u.problemRepo.FindAll(
		ctx,
		strings.TrimSpace(search),
		page,
		limit,
	)
}

func (u *ProblemUsecase) Create(ctx context.Context, in model.ProblemInput) (*model.Problem, error) {
	name, err := validateProblemName(in.Name)
	if err != nil {
		return nil, err
	}

	exists, err := u.problemRepo.ExistsByName(
		ctx,
		name,
		0,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New(
			"permasalahan tersebut sudah terdaftar",
		)
	}

	problem := &model.Problem{
		Name: name,
	}

	if err := u.problemRepo.Create(ctx, problem); err != nil {
		return nil, err
	}

	return problem, nil
}

func (u *ProblemUsecase) Update(ctx context.Context, id int64, in model.ProblemInput) (*model.Problem, error) {
	if id <= 0 {
		return nil, errors.New(
			"ID permasalahan tidak valid",
		)
	}

	name, err := validateProblemName(in.Name)
	if err != nil {
		return nil, err
	}

	exists, err := u.problemRepo.ExistsByName(
		ctx,
		name,
		id,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New(
			"permasalahan tersebut sudah terdaftar",
		)
	}

	problem := &model.Problem{
		ID:   id,
		Name: name,
	}

	if err := u.problemRepo.Update(ctx, problem); err != nil {
		return nil, err
	}

	return u.problemRepo.FindByID(ctx, id)
}

func (u *ProblemUsecase) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New(
			"ID permasalahan tidak valid",
		)
	}

	return u.problemRepo.Delete(ctx, id)
}
