package usecase

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"

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
	if err := validate.Struct(in); err != nil {
		return "", err
	}

	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		return "", errors.New("email or password is incorrect")
	}

	if !helper.CheckPasswordHash(in.Password, user.Password) {
		return "", errors.New("email or password is incorrect")
	}

	return helper.GenerateToken(*user)
}

func (u *UserUsecase) Create(ctx context.Context, in model.CreateUserInput) (string, error) {
	if err := validate.Struct(in); err != nil {
		return "", err
	}

	hashed, err := helper.HashRequestPassword(in.Password)
	if err != nil {
		return "", err
	}

	newUser, err := u.userRepo.Create(ctx, model.User{
		Name:     in.Name,
		Email:    in.Email,
		Password: hashed,
		RoleID:   in.RoleID,
		IsActive: true,
	})

	if err != nil {
		return "", err
	}

	return helper.GenerateToken(*newUser)
}

func (u *UserUsecase) FindAll(ctx context.Context, filter model.User) ([]*model.User, error) {
	return u.userRepo.FindAll(ctx, filter)
}

func (u *UserUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	return u.userRepo.FindByID(ctx, id)
}

func (u *UserUsecase) Update(ctx context.Context, id int64, in model.UpdateUserInput) error {
	if err := validate.Struct(in); err != nil {
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

	return u.userRepo.Update(ctx, *user)
}

func (u *UserUsecase) Delete(ctx context.Context, id int64) error {
	return u.userRepo.Delete(ctx, id)
}
