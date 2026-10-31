package usecase

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

var appSettingColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
var appSettingPhonePattern = regexp.MustCompile(`^[0-9]{8,15}$`)

type AppSettingUsecase struct {
	repo model.IAppSettingRepository
}

func validatePriorityMinutes(priority string, minutes int) error {
	if minutes < 1 || minutes > 525600 {
		return errors.New(
			priority +
				" SLA harus diantara 1 menit sampai 525600 menit",
		)
	}

	return nil
}

func NewAppSettingUsecase(repo model.IAppSettingRepository) model.IAppSettingUsecase {
	return &AppSettingUsecase{repo: repo}
}

func (u *AppSettingUsecase) Get(ctx context.Context) (*model.AppSetting, error) {
	return u.repo.Get(ctx)
}

func (u *AppSettingUsecase) Update(ctx context.Context, setting model.AppSetting) (*model.AppSetting, error) {
	setting.ID = 1
	setting.AppName = strings.TrimSpace(setting.AppName)
	setting.LoginTitle = strings.TrimSpace(setting.LoginTitle)
	setting.Tagline = strings.TrimSpace(setting.Tagline)
	setting.WhatsAppNumber = strings.TrimSpace(setting.WhatsAppNumber)
	setting.WhatsAppDisplayNumber = strings.TrimSpace(setting.WhatsAppDisplayNumber)
	setting.WhatsAppMessage = strings.TrimSpace(setting.WhatsAppMessage)
	setting.PrimaryColor = strings.TrimSpace(setting.PrimaryColor)
	setting.SidebarTitle = strings.TrimSpace(setting.SidebarTitle)
	setting.FooterText = strings.TrimSpace(setting.FooterText)
	setting.AppVersion = strings.TrimSpace(setting.AppVersion)

	if setting.AppName == "" {
		return nil, errors.New("nama aplikasi wajib diisi")
	}
	if len([]rune(setting.AppName)) > 100 {
		return nil, errors.New("nama aplikasi maksimal 100 karakter")
	}
	if setting.LoginTitle == "" {
		return nil, errors.New("judul halaman login wajib diisi")
	}
	if len([]rune(setting.LoginTitle)) > 100 {
		return nil, errors.New("judul halaman login maksimal 100 karakter")
	}
	if len([]rune(setting.Tagline)) > 200 {
		return nil, errors.New("tagline maksimal 200 karakter")
	}
	if len([]rune(setting.WhatsAppMessage)) > 4000 {
		return nil, errors.New("pesan WhatsApp maksimal 4000 karakter")
	}
	if setting.PrimaryColor == "" {
		setting.PrimaryColor = "#EA580C"
	}
	if !appSettingColorPattern.MatchString(setting.PrimaryColor) {
		return nil, errors.New("warna utama harus menggunakan format HEX, contoh #EA580C")
	}
	if setting.WhatsAppNumber != "" && !appSettingPhonePattern.MatchString(setting.WhatsAppNumber) {
		return nil, errors.New("nomor WhatsApp harus berisi 8 sampai 15 digit termasuk kode negara")
	}
	if len([]rune(setting.WhatsAppDisplayNumber)) > 40 {
		return nil, errors.New("nomor WhatsApp yang ditampilkan maksimal 40 karakter")
	}
	if len([]rune(setting.SidebarTitle)) > 100 {
		return nil, errors.New("judul Sidebar maksimal 100 karakter")
	}
	if len([]rune(setting.AppVersion)) > 30 {
		return nil, errors.New("versi aplikasi maksimal 30 karakter")
	}

	if err := validatePriorityMinutes(
		"URGENT",
		setting.PriorityUrgentMinutes,
	); err != nil {
		return nil, err
	}

	if err := validatePriorityMinutes(
		"HIGH",
		setting.PriorityHighMinutes,
	); err != nil {
		return nil, err
	}

	if err := validatePriorityMinutes(
		"MEDIUM",
		setting.PriorityMediumMinutes,
	); err != nil {
		return nil, err
	}

	if err := validatePriorityMinutes(
		"LOW",
		setting.PriorityLowMinutes,
	); err != nil {
		return nil, err
	}

	if err := validateLoginSessionMinutes(
		setting.LoginSessionMinutes,
	); err != nil {
		return nil, err
	}

	if err := validateStaffResponseMinutes(
		setting.StaffResponseSLAMinutes,
	); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, setting); err != nil {
		return nil, err
	}
	return u.repo.Get(ctx)
}

func validateLoginSessionMinutes(minutes int) error {
	if minutes < 5 || minutes > 43200 {
		return errors.New(
			"durasi login harus antara 5 sampai 43200 menit",
		)
	}

	return nil
}

func validateStaffResponseMinutes(minutes int) error {
	if minutes < 1 || minutes > 525600 {
		return errors.New(
			"SLA respons tiket harus antara 1 sampai 525600 menit",
		)
	}

	return nil
}
