package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type ProjectUsecase struct {
	projectRepo model.IProjectRepository
}

func NewProjectUsecase(projectRepo model.IProjectRepository) model.IProjectUsecase {
	return &ProjectUsecase{
		projectRepo: projectRepo,
	}
}

func (u *ProjectUsecase) Create(ctx context.Context, in model.CreateProjectInput) (*model.Project, error) {
	log := logrus.WithFields(logrus.Fields{
		"in": in,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	project := model.Project{
		Name: in.Name,
	}

	created, err := u.projectRepo.Create(ctx, project)
	if err != nil {
		log.Error("Failed to create project: ", err)
		return nil, err
	}

	return created, nil
}

func (u *ProjectUsecase) FindAll(ctx context.Context, filter model.Project) ([]*model.Project, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	projects, err := u.projectRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch projects: ", err)
		return nil, err
	}

	return projects, nil
}

func (u *ProjectUsecase) FindByID(ctx context.Context, id int64) (*model.Project, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	project, err := u.projectRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find project: ", err)
		return nil, err
	}

	return project, nil
}

func (u *ProjectUsecase) Update(ctx context.Context, id int64, in model.UpdateProjectInput) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	project, err := u.projectRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	project.Name = in.Name

	if err := u.projectRepo.Update(ctx, *project); err != nil {
		log.Error("Failed to update project: ", err)
		return err
	}

	return nil
}

func (u *ProjectUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	project, err := u.projectRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find project for deletion: ", err)
		return err
	}

	if project == nil {
		return errors.New("project not found")
	}

	return u.projectRepo.Delete(ctx, id)
}
