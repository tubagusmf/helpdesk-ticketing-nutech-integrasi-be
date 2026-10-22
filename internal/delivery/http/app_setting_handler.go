package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

const (
	appSettingUploadURLPrefix = "/uploads/app-settings/"
	maxAppSettingRequestSize  = 16 << 20
	maxBrandImageSize         = 5 << 20
	maxFaviconSize            = 1 << 20
)

type AppSettingHandler struct {
	usecase model.IAppSettingUsecase
}

type savedAppSettingFile struct {
	URL  string
	Path string
}

func NewAppSettingHandler(e *echo.Echo, appSettingUsecase model.IAppSettingUsecase) {
	handler := &AppSettingHandler{usecase: appSettingUsecase}
	group := e.Group("/v1/app-settings")

	group.GET("/public", handler.GetPublic)
	group.GET("", handler.Get, AuthMiddleware)
	group.PUT("", handler.Update, AuthMiddleware)
}

func (h *AppSettingHandler) GetPublic(c echo.Context) error {
	return h.getSettings(c)
}

func (h *AppSettingHandler) Get(c echo.Context) error {
	if err := requireAppSettingAdministrator(c); err != nil {
		return err
	}
	return h.getSettings(c)
}

func (h *AppSettingHandler) getSettings(c echo.Context) error {
	setting, err := h.usecase.Get(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve application settings")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "app settings fetched successfully",
		"data":    setting,
	})
}

func (h *AppSettingHandler) Update(c echo.Context) error {
	if err := requireAppSettingAdministrator(c); err != nil {
		return err
	}

	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxAppSettingRequestSize)
	if err := c.Request().ParseMultipartForm(2 << 20); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "requests must be multipart/form-data and file size must not exceed the limit")
	}
	if c.Request().MultipartForm == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "multipart/form-data is required")
	}

	setting, err := h.usecase.Get(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve application settings")
	}
	oldSetting := *setting
	formValues := c.Request().MultipartForm.Value
	readField := func(key, previous string) string {
		values, exists := formValues[key]
		if !exists || len(values) == 0 {
			return previous
		}
		return values[0]
	}

	setting.AppName = readField("app_name", setting.AppName)
	setting.LoginTitle = readField("login_title", setting.LoginTitle)
	setting.Tagline = readField("tagline", setting.Tagline)
	setting.WhatsAppNumber = readField("whatsapp_number", setting.WhatsAppNumber)
	setting.WhatsAppDisplayNumber = readField("whatsapp_display_number", setting.WhatsAppDisplayNumber)
	setting.WhatsAppMessage = readField("whatsapp_message", setting.WhatsAppMessage)
	setting.PrimaryColor = readField("primary_color", setting.PrimaryColor)
	setting.SidebarTitle = readField("sidebar_title", setting.SidebarTitle)
	setting.FooterText = readField("footer_text", setting.FooterText)
	setting.AppVersion = readField("app_version", setting.AppVersion)

	savedFiles := make([]savedAppSettingFile, 0, 3)
	persisted := false
	defer func() {
		if !persisted {
			for _, file := range savedFiles {
				_ = os.Remove(file.Path)
			}
		}
	}()

	logoURL, logoFile, uploadErr := saveAppSettingImage(c, "logo", "logo", maxBrandImageSize, false)
	if uploadErr != nil {
		return uploadErr
	}
	if logoFile != nil {
		savedFiles = append(savedFiles, *logoFile)
		setting.LogoURL = logoURL
	}

	iconURL, iconFile, uploadErr := saveAppSettingImage(c, "login_icon", "login-icon", maxBrandImageSize, false)
	if uploadErr != nil {
		return uploadErr
	}
	if iconFile != nil {
		savedFiles = append(savedFiles, *iconFile)
		setting.LoginIconURL = iconURL
	}

	faviconURL, faviconFile, uploadErr := saveAppSettingImage(c, "favicon", "favicon", maxFaviconSize, true)
	if uploadErr != nil {
		return uploadErr
	}
	if faviconFile != nil {
		savedFiles = append(savedFiles, *faviconFile)
		setting.FaviconURL = faviconURL
	}

	updated, err := h.usecase.Update(c.Request().Context(), *setting)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	persisted = true

	newURLs := map[string]bool{
		updated.LogoURL:      true,
		updated.LoginIconURL: true,
		updated.FaviconURL:   true,
	}
	for _, oldURL := range []string{oldSetting.LogoURL, oldSetting.LoginIconURL, oldSetting.FaviconURL} {
		if oldURL != "" && oldURL != updated.LogoURL && oldURL != updated.LoginIconURL && oldURL != updated.FaviconURL && !newURLs[oldURL] {
			removeAppSettingUpload(oldURL)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "app settings updated successfully",
		"data":    updated,
	})
}

func requireAppSettingAdministrator(c echo.Context) error {
	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}
	claim, ok := claimValue.(*model.CustomClaims)
	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user claims")
	}
	if claim.RoleID != 1 {
		return echo.NewHTTPError(http.StatusForbidden, "only Administrator can manage application settings")
	}
	return nil
}

func saveAppSettingImage(c echo.Context, fieldName string, filenamePrefix string, maxSize int64, allowICO bool) (string, *savedAppSettingFile, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil, nil
		}
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to read file %s", fieldName))
	}
	if fileHeader.Size <= 0 {
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file %s empty", fieldName))
	}
	if fileHeader.Size > maxSize {
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file size %s exceeds limit", fieldName))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !validAppSettingImage(fileHeader, ext, allowICO) {
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file format %s is invalid", fieldName))
	}

	source, err := fileHeader.Open()
	if err != nil {
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to open file %s", fieldName))
	}
	defer source.Close()

	directory := filepath.Join(helper.UploadDir(), "app-settings")
	if err := os.MkdirAll(directory, 0750); err != nil {
		return "", nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create application settings upload directory")
	}

	filename := fmt.Sprintf("%s-%d%s", filenamePrefix, time.Now().UnixNano(), ext)
	fullPath := filepath.Join(directory, filename)
	destination, err := os.OpenFile(fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return "", nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to save application settings image")
	}

	written, copyErr := io.Copy(destination, io.LimitReader(source, maxSize+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written <= 0 || written > maxSize {
		_ = os.Remove(fullPath)
		return "", nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("failed to save file %s or file size exceeds limit", fieldName))
	}

	saved := &savedAppSettingFile{
		URL:  appSettingUploadURLPrefix + filename,
		Path: fullPath,
	}
	return saved.URL, saved, nil
}

func validAppSettingImage(fileHeader *multipart.FileHeader, ext string, allowICO bool) bool {
	allowed := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".webp": true,
	}
	if allowICO {
		allowed[".ico"] = true
	}
	if !allowed[ext] {
		return false
	}

	file, err := fileHeader.Open()
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 16)
	n, readErr := file.Read(header)
	if readErr != nil && readErr != io.EOF {
		return false
	}
	header = header[:n]

	switch ext {
	case ".png":
		return bytes.HasPrefix(header, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	case ".jpg", ".jpeg":
		return len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF
	case ".webp":
		return len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP"
	case ".ico":
		return allowICO && len(header) >= 4 && header[0] == 0x00 && header[1] == 0x00 && header[2] == 0x01 && header[3] == 0x00
	default:
		return false
	}
}

func removeAppSettingUpload(url string) {
	if !strings.HasPrefix(url, appSettingUploadURLPrefix) {
		return
	}
	filename := strings.TrimPrefix(url, appSettingUploadURLPrefix)
	if filename == "" || filepath.Base(filename) != filename || strings.Contains(filename, "..") {
		return
	}
	_ = os.Remove(filepath.Join(helper.UploadDir(), "app-settings", filename))
}
