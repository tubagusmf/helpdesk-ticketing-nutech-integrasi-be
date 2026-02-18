package usecase

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

var validate = validator.New()

type UserUsecase struct {
	userRepo model.IUserRepository
}

func NewUserUsecase(userRepo model.IUserRepository) model.IUserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (u *UserUsecase) Login(ctx context.Context, in model.LoginInput) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"email": in.Email,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error", err)
		return "", err
	}

	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		return "", errors.New("email or password is incorrect")
	}

	if !helper.CheckPasswordHash(in.Password, user.Password) {
		return "", errors.New("email or password is incorrect")
	}

	token, err := helper.GenerateToken(*user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserUsecase) Create(ctx context.Context, in model.CreateUserInput) (string, error) {
	log := logrus.WithFields(logrus.Fields{
		"in": in,
	})

	if err := validate.Struct(in); err != nil {
		return "", err
	}

	hashed, err := helper.HashRequestPassword(in.Password)
	if err != nil {
		log.Error(err)
		return "", err
	}

	var projects []model.Project
	for _, p := range in.Projects {
		projects = append(projects, model.Project{
			ID: p.ID,
		})
	}

	newUser, err := u.userRepo.Create(ctx, model.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: hashed,
		RoleID:   in.RoleID,
		IsActive: true,
		Projects: projects,
	})

	if err != nil {
		log.Error(err)
		return "", err
	}

	logrus.Infof("Projects count: %d", len(newUser.Projects))

	accessToken, err := helper.GenerateToken(*newUser)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (u *UserUsecase) FindAll(ctx context.Context, filter model.User) ([]*model.User, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	users, err := u.userRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return users, nil
}

func (u *UserUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) Update(ctx context.Context, id int64, in model.UpdateUserInput) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error", err)
		return err
	}

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	hashed, err := helper.HashRequestPassword(in.Password)
	if err != nil {
		return err
	}

	user.Name = in.Name
	user.Email = in.Email
	user.Password = hashed
	user.RoleID = in.RoleID

	var projects []model.Project
	for _, p := range in.Projects {
		projects = append(projects, model.Project{
			ID: p.ID,
		})
	}

	user.Projects = projects

	logrus.Infof("Projects count: %d", len(user.Projects))

	return u.userRepo.Update(ctx, *user)
}

func (u *UserUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find user for deletion: ", err)
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	return u.userRepo.Delete(ctx, id)
}
