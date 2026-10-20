package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) model.IUserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user model.User) (*model.User, error) {
	tx := r.db.WithContext(ctx).Begin()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := tx.Omit("Projects.*").Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, project := range user.Projects {
		userProject := map[string]interface{}{
			"user_id":    user.ID,
			"project_id": project.ID,
			"created_at": now,
			"updated_at": now,
		}

		if err := tx.Table("user_projects").
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "user_id"}, {Name: "project_id"}},
				DoNothing: true,
			}).
			Create(userProject).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return &user, tx.Commit().Error
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Projects").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Role").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) FindAll(ctx context.Context, filter model.User, page int, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("users.deleted_at IS NULL").
		Preload("Role").
		Preload("Projects")

	if filter.Name != "" {
		search := "%" + filter.Name + "%"

		query = query.Where(`
		(
			users.name ILIKE ?
			OR users.email ILIKE ?
			OR EXISTS (
				SELECT 1
				FROM user_projects
				JOIN projects ON projects.id = user_projects.project_id
				WHERE user_projects.user_id = users.id
				AND projects.name ILIKE ?
			)
			OR EXISTS (
				SELECT 1
				FROM roles
				WHERE roles.id = users.role_id
				AND roles.name ILIKE ?
			)
		)
		`, search, search, search, search)
	}

	if filter.RoleID != 0 {
		query = query.Where("users.role_id = ?", filter.RoleID)
	}

	if filter.IsActive {
		query = query.Where("users.is_active = ?", true)
	}

	if filter.ProjectID != 0 {
		query = query.Where(`
		EXISTS (
			SELECT 1
			FROM user_projects
			WHERE user_projects.user_id = users.id
			AND user_projects.project_id = ?
		)
	`, filter.ProjectID)
	}

	if filter.IsOnline {
		query = query.Where(`
			users.last_seen > NOW() - INTERVAL '1 minute'
		`)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Limit(limit).
		Offset(offset).
		Order("users.id DESC").
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepo) Update(ctx context.Context, user model.User) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	var codeNameValue interface{}

	if user.CodeName != nil {
		value := strings.TrimSpace(*user.CodeName)

		if value != "" {
			codeNameValue = value
		}
	}

	if err := tx.Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(map[string]interface{}{
			"name":       user.Name,
			"code_name":  codeNameValue,
			"email":      user.Email,
			"password":   user.Password,
			"role_id":    user.RoleID,
			"is_active":  user.IsActive,
			"updated_at": now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Table("user_projects").
		Where("user_id = ?", user.ID).
		Delete(nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, project := range user.Projects {
		userProject := map[string]interface{}{
			"user_id":    user.ID,
			"project_id": project.ID,
			"created_at": now,
			"updated_at": now,
		}

		if err := tx.Table("user_projects").
			Create(userProject).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	tx := r.db.WithContext(ctx).Begin()

	now := time.Now()

	if err := tx.Model(&model.User{}).
		Where("id = ?", id).
		Update("deleted_at", now).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Table("user_projects").
		Where("user_id = ?", id).
		Delete(nil).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *UserRepo) UpdateOnlineStatus(ctx context.Context, userID int64, isOnline bool) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	now := time.Now()

	if err := tx.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_online":  isOnline,
			"last_seen":  now,
			"updated_at": now,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// if isOnline {
	// 	var session model.UserOfflineSession

	// 	err := tx.
	// 		Where("user_id = ? AND online_at IS NULL", userID).
	// 		Order("offline_at DESC").
	// 		First(&session).Error

	// 	if err == nil {
	// 		duration := int64(now.Sub(session.OfflineAt).Seconds())
	// 		if duration < 0 {
	// 			duration = 0
	// 		}

	// 		if err := tx.Model(&model.UserOfflineSession{}).
	// 			Where("id = ?", session.ID).
	// 			Updates(map[string]interface{}{
	// 				"online_at":        now,
	// 				"duration_seconds": duration,
	// 				"updated_at":       now,
	// 			}).Error; err != nil {
	// 			tx.Rollback()
	// 			return err
	// 		}
	// 	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 		tx.Rollback()
	// 		return err
	// 	}
	// }

	return tx.Commit().Error
}

func (r *UserRepo) UpdateLastSeen(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_seen", time.Now()).Error
}

// func (r *UserRepo) ChangeOnlineStatus(ctx context.Context, userID int64, isOnline bool, justification string) (*model.UserOfflineSession, error) {
// 	tx := r.db.WithContext(ctx).Begin()
// 	if tx.Error != nil {
// 		return nil, tx.Error
// 	}

// 	var user model.User

// 	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
// 		Where("id = ? AND deleted_at IS NULL", userID).
// 		First(&user).Error
// 	if err != nil {
// 		tx.Rollback()

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, errors.New("user not found")
// 		}

// 		return nil, err
// 	}

// 	now := time.Now()

// 	// OFFLINE -> ONLINE
// 	if isOnline {
// 		var session model.UserOfflineSession
// 		var completed *model.UserOfflineSession

// 		err := tx.
// 			Where("user_id = ? AND online_at IS NULL", userID).
// 			Order("offline_at DESC").
// 			First(&session).Error

// 		if err == nil {
// 			duration := int64(now.Sub(session.OfflineAt).Seconds())
// 			if duration < 0 {
// 				duration = 0
// 			}

// 			if err := tx.Model(&model.UserOfflineSession{}).
// 				Where("id = ?", session.ID).
// 				Updates(map[string]interface{}{
// 					"online_at":        now,
// 					"duration_seconds": duration,
// 					"updated_at":       now,
// 				}).Error; err != nil {
// 				tx.Rollback()
// 				return nil, err
// 			}

// 			session.OnlineAt = &now
// 			session.DurationSeconds = duration
// 			session.UpdatedAt = now
// 			completed = &session
// 		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			tx.Rollback()
// 			return nil, err
// 		}

// 		if err := tx.Model(&model.User{}).
// 			Where("id = ?", userID).
// 			Updates(map[string]interface{}{
// 				"is_online":  true,
// 				"last_seen":  now,
// 				"updated_at": now,
// 			}).Error; err != nil {
// 			tx.Rollback()
// 			return nil, err
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			return nil, err
// 		}

// 		return completed, nil
// 	}

// 	if !user.IsOnline {
// 		var active model.UserOfflineSession

// 		err := tx.
// 			Where("user_id = ? AND online_at IS NULL", userID).
// 			Order("offline_at DESC").
// 			First(&active).Error

// 		if err == nil {
// 			if err := tx.Commit().Error; err != nil {
// 				return nil, err
// 			}

// 			return &active, nil
// 		}

// 		tx.Rollback()

// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, errors.New("user is already offline")
// 		}

// 		return nil, err
// 	}

// 	justification = strings.TrimSpace(justification)
// 	if justification == "" {
// 		tx.Rollback()
// 		return nil, errors.New("justifikasi wajib diisi")
// 	}

// 	session := model.UserOfflineSession{
// 		UserID:        userID,
// 		Justification: justification,
// 		OfflineAt:     now,
// 	}

// 	if err := tx.Create(&session).Error; err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	if err := tx.Model(&model.User{}).
// 		Where("id = ?", userID).
// 		Updates(map[string]interface{}{
// 			"is_online":  false,
// 			"last_seen":  now,
// 			"updated_at": now,
// 		}).Error; err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		return nil, err
// 	}

// 	return &session, nil
// }

// func (r *UserRepo) GetActiveOfflineSession(ctx context.Context, userID int64) (*model.UserOfflineSession, error) {
// 	var session model.UserOfflineSession

// 	err := r.db.WithContext(ctx).
// 		Where("user_id = ? AND online_at IS NULL", userID).
// 		Order("offline_at DESC").
// 		First(&session).Error

// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		return nil, nil
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &session, nil
// }
