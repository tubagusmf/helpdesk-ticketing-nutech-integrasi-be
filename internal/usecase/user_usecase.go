package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

var validate = validator.New()

type UserUsecase struct {
	userRepo       model.IUserRepository
	projectRepo    model.IProjectRepository
	appSettingRepo model.IAppSettingRepository
}

func NewUserUsecase(userRepo model.IUserRepository, projectRepo model.IProjectRepository, appSettingRepo model.IAppSettingRepository) model.IUserUsecase {
	return &UserUsecase{
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		appSettingRepo: appSettingRepo,
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

	if !helper.VerifyLoginCaptcha(in.CaptchaID, in.CaptchaCode) {
		return "", errors.New(
			"Kode CAPTCHA Salah",
		)
	}

	user, err := u.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		return "", errors.New("email or password is incorrect")
	}

	if !helper.CheckPasswordHash(in.Password, user.Password) {
		return "", errors.New("email or password is incorrect")
	}

	// err = u.userRepo.UpdateOnlineStatus(ctx, user.ID, true)
	// if err != nil {
	// 	log.Error("failed update online status:", err)
	// }

	token, err := u.generateAccessToken(ctx, *user)
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

	codeName := strings.TrimSpace(in.CodeName)

	if in.RoleID == 2 {
		if codeName == "" {
			return "", errors.New(
				"Code Name wajib diisi untuk Staff / CCIT",
			)
		}

		if len([]rune(codeName)) > 20 {
			return "", errors.New(
				"Code Name maksimal 20 karakter",
			)
		}
	}

	var codeNamePtr *string

	if in.RoleID == 2 {
		codeNamePtr = &codeName
	}

	var projects []model.Project
	for _, p := range in.Projects {
		projects = append(projects, model.Project{
			ID: p.ID,
		})
	}

	newUser, err := u.userRepo.Create(ctx, model.User{
		Name:     in.Name,
		CodeName: codeNamePtr,
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

	accessToken, err := u.generateAccessToken(ctx, *newUser)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (u *UserUsecase) FindAll(ctx context.Context, filter model.User, page int, limit int) ([]*model.User, int64, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	users, total, err := u.userRepo.FindAll(ctx, filter, page, limit)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}

	return users, total, nil
}

func (u *UserUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Administrator can see all projects
	if user.Role.Name == "ADMINISTRATOR" {
		projects, _, err := u.projectRepo.FindAll(ctx, model.Project{}, 1, 1000)
		if err != nil {
			return nil, err
		}

		var allProjects []model.Project
		for _, p := range projects {
			allProjects = append(allProjects, *p)
		}

		user.Projects = allProjects
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

	user.Name = in.Name
	user.Email = in.Email
	user.RoleID = in.RoleID

	codeName := strings.TrimSpace(in.CodeName)

	if len([]rune(codeName)) > 20 {
		return errors.New("Code Name maksimal 20 karakter")
	}

	if user.RoleID == 2 && codeName != "" {
		user.CodeName = &codeName
	} else {
		user.CodeName = nil
	}

	if in.Password != "" {
		hashed, err := helper.HashRequestPassword(in.Password)
		if err != nil {
			return err
		}

		user.Password = hashed
	}

	var projects []model.Project
	for _, p := range in.Projects {
		projects = append(projects, model.Project{
			ID: p.ID,
		})
	}

	user.Projects = projects

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

func (u *UserUsecase) UpdateOnlineStatus(ctx context.Context, userID int64, isOnline bool) error {
	return u.userRepo.UpdateOnlineStatus(ctx, userID, isOnline)
}

func (u *UserUsecase) UpdateHeartbeat(ctx context.Context, userID int64) error {
	return u.userRepo.UpdateHeartbeat(ctx, userID)
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, userID int64, in model.UpdateProfileInput) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Name = in.Name

	if in.NewPassword != "" {

		if in.CurrentPassword == "" {
			return errors.New("current password is required")
		}

		if !helper.CheckPasswordHash(in.CurrentPassword, user.Password) {
			return errors.New("current password is incorrect")
		}

		hashed, err := helper.HashRequestPassword(in.NewPassword)
		if err != nil {
			return err
		}

		user.Password = hashed
	}

	return u.userRepo.Update(ctx, *user)
}

func (u *UserUsecase) MarkStaleUsersOffline(ctx context.Context, cutoff time.Time) error {
	return u.userRepo.MarkStaleUsersOffline(ctx, cutoff)
}

func (u *UserUsecase) generateAccessToken(ctx context.Context, user model.User) (string, error) {
	settings, err := u.appSettingRepo.Get(ctx)
	if err != nil {
		return "", err
	}

	loginMinutes := settings.LoginSessionMinutes

	if loginMinutes < 5 || loginMinutes > 43200 {
		return "", errors.New(
			"konfigurasi durasi login tidak valid",
		)
	}

	ttl := time.Duration(loginMinutes) * time.Minute

	return helper.GenerateTokenWithTTL(
		user,
		ttl,
	)
}

// func (u *UserUsecase) ChangeOnlineStatus(ctx context.Context, userID int64, isOnline bool, justification string) (*model.UserOfflineSession, error) {
// 	justification = strings.TrimSpace(justification)

// 	if !isOnline {
// 		if justification == "" {
// 			return nil, errors.New("justifikasi wajib diisi")
// 		}

// 		if len([]rune(justification)) > 1000 {
// 			return nil, errors.New(
// 				"justifikasi maksimal 1000 karakter",
// 			)
// 		}
// 	}

// 	return u.userRepo.ChangeOnlineStatus(
// 		ctx,
// 		userID,
// 		isOnline,
// 		justification,
// 	)
// }

// func (u *UserUsecase) GetActiveOfflineSession(ctx context.Context, userID int64) (*model.UserOfflineSession, error) {
// 	return u.userRepo.GetActiveOfflineSession(ctx, userID)
// }
