package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppSettingRepo struct {
	db *gorm.DB
}

func NewAppSettingRepo(db *gorm.DB) model.IAppSettingRepository {
	return &AppSettingRepo{db: db}
}

func defaultAppSetting() model.AppSetting {
	return model.AppSetting{
		ID:                      1,
		AppName:                 "Helpdesk Center",
		LoginTitle:              "Helpdesk Center",
		Tagline:                 "Professional Ticketing System",
		WhatsAppNumber:          "6281119521124",
		WhatsAppDisplayNumber:   "081119521124",
		WhatsAppMessage:         "Halo Admin,\n\nMohon untuk mendaftar akun Helpdesk dengan mengisi data berikut:\n\nNama Lengkap:\nEmail:\nProject:\n\nTerima kasih.",
		PrimaryColor:            "#EA580C",
		SidebarTitle:            "Helpdesk Center",
		FooterText:              "© 2026 Helpdesk CCIT Nutech Integrasi",
		AppVersion:              "Version 1.0.0",
		PriorityUrgentMinutes:   15,
		PriorityHighMinutes:     60,
		PriorityMediumMinutes:   120,
		PriorityLowMinutes:      240,
		LoginSessionMinutes:     600,
		StaffResponseSLAMinutes: 3,
	}
}

func (r *AppSettingRepo) Get(ctx context.Context) (*model.AppSetting, error) {
	var setting model.AppSetting
	err := r.db.WithContext(ctx).First(&setting, "id = ?", 1).Error
	if err == nil {
		return &setting, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	setting = defaultAppSetting()
	if createErr := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&setting).Error; createErr != nil {
		return nil, createErr
	}

	if err := r.db.WithContext(ctx).First(&setting, "id = ?", 1).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *AppSettingRepo) Update(ctx context.Context, setting model.AppSetting) error {
	updates := map[string]interface{}{
		"app_name":                   setting.AppName,
		"login_title":                setting.LoginTitle,
		"tagline":                    setting.Tagline,
		"logo_url":                   setting.LogoURL,
		"login_icon_url":             setting.LoginIconURL,
		"favicon_url":                setting.FaviconURL,
		"whatsapp_number":            setting.WhatsAppNumber,
		"whatsapp_display_number":    setting.WhatsAppDisplayNumber,
		"whatsapp_message":           setting.WhatsAppMessage,
		"primary_color":              setting.PrimaryColor,
		"sidebar_title":              setting.SidebarTitle,
		"footer_text":                setting.FooterText,
		"app_version":                setting.AppVersion,
		"priority_urgent_minutes":    setting.PriorityUrgentMinutes,
		"priority_high_minutes":      setting.PriorityHighMinutes,
		"priority_medium_minutes":    setting.PriorityMediumMinutes,
		"priority_low_minutes":       setting.PriorityLowMinutes,
		"login_session_minutes":      setting.LoginSessionMinutes,
		"staff_response_sla_minutes": setting.StaffResponseSLAMinutes,
		"updated_at":                 time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&model.AppSetting{}).
		Where("id = ?", 1).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if _, err := r.Get(ctx); err != nil {
			return err
		}
		result = r.db.WithContext(ctx).
			Model(&model.AppSetting{}).
			Where("id = ?", 1).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
