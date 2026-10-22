package model

import (
	"context"
	"time"
)

type AppSetting struct {
	ID                    int64     `json:"id"`
	AppName               string    `json:"app_name"`
	LoginTitle            string    `json:"login_title"`
	Tagline               string    `json:"tagline"`
	LogoURL               string    `json:"logo_url"`
	LoginIconURL          string    `json:"login_icon_url"`
	FaviconURL            string    `json:"favicon_url"`
	WhatsAppNumber        string    `json:"whatsapp_number"`
	WhatsAppDisplayNumber string    `json:"whatsapp_display_number"`
	WhatsAppMessage       string    `json:"whatsapp_message"`
	PrimaryColor          string    `json:"primary_color"`
	SidebarTitle          string    `json:"sidebar_title"`
	FooterText            string    `json:"footer_text"`
	AppVersion            string    `json:"app_version"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (AppSetting) TableName() string {
	return "app_settings"
}

type IAppSettingRepository interface {
	Get(ctx context.Context) (*AppSetting, error)
	Update(ctx context.Context, setting AppSetting) error
}

type IAppSettingUsecase interface {
	Get(ctx context.Context) (*AppSetting, error)
	Update(ctx context.Context, setting AppSetting) (*AppSetting, error)
}
