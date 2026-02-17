package usecase

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type RoleUsecase struct {
	roleRepo model.IRoleRepository
}

func NewRoleUsecase(roleRepo model.IRoleRepository) model.IRoleUsecase {
	return &RoleUsecase{
		roleRepo: roleRepo,
	}
}

func (u *RoleUsecase) Create(ctx context.Context, in model.CreateRoleInput) (*model.Role, error) {
	log := logrus.WithFields(logrus.Fields{
		"in": in,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	role := model.Role{
		Name:      in.Name,
		Privilege: in.Privilege,
	}

	created, err := u.roleRepo.Create(ctx, role)
	if err != nil {
		log.Error("Failed to create role: ", err)
		return nil, err
	}

	return created, nil
}

func (u *RoleUsecase) FindAll(ctx context.Context, filter model.Role) ([]*model.Role, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	roles, err := u.roleRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch roles: ", err)
		return nil, err
	}

	return roles, nil
}

func (u *RoleUsecase) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	role, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find role: ", err)
		return nil, err
	}

	return role, nil
}

func (u *RoleUsecase) Update(ctx context.Context, id int64, in model.UpdateRoleInput) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	role, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Role not found: ", err)
		return err
	}

	role.Name = in.Name
	role.Privilege = in.Privilege

	if err := u.roleRepo.Update(ctx, *role); err != nil {
		log.Error("Failed to update role: ", err)
		return err
	}

	return nil
}

func (u *RoleUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	role, err := u.roleRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find role for deletion: ", err)
		return err
	}

	if role == nil {
		return errors.New("role not found")
	}

	if err := u.roleRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete role: ", err)
		return err
	}

	return nil
}
