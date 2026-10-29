package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type FCMTokenHandler struct {
	usecase model.IUserFCMTokenUsecase
}

func NewFCMTokenHandler(e *echo.Echo, usecase model.IUserFCMTokenUsecase) {
	handler := &FCMTokenHandler{usecase: usecase}

	group := e.Group("/v1/fcm")
	group.POST("/tokens", handler.Register, AuthMiddleware)
	group.DELETE("/tokens", handler.Unregister, AuthMiddleware)
}

func getFCMAuthenticatedUserID(c echo.Context) (int64, error) {
	claim, ok := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	if !ok || claim == nil || claim.UserID <= 0 {
		return 0, echo.NewHTTPError(
			http.StatusUnauthorized,
			"unauthorized",
		)
	}

	return claim.UserID, nil
}

type FCMTokenRequest struct {
	Token string `json:"token"`
}

func (h *FCMTokenHandler) Register(c echo.Context) error {
	userID, err := getFCMAuthenticatedUserID(c)
	if err != nil {
		return err
	}

	var req FCMTokenRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"request body tidak valid",
		)
	}

	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" || len(req.Token) > 4096 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"FCM token wajib diisi dan harus valid",
		)
	}

	if err := h.usecase.RegisterToken(
		c.Request().Context(),
		userID,
		req.Token,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal mendaftarkan perangkat",
		)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "perangkat berhasil didaftarkan",
	})
}

func (h *FCMTokenHandler) Unregister(c echo.Context) error {
	userID, err := getFCMAuthenticatedUserID(c)
	if err != nil {
		return err
	}

	var req FCMTokenRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"request body tidak valid",
		)
	}

	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" || len(req.Token) > 4096 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"FCM token wajib diisi dan harus valid",
		)
	}

	if err := h.usecase.UnregisterToken(
		c.Request().Context(),
		userID,
		req.Token,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal menghapus perangkat",
		)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "perangkat berhasil dihapus",
	})
}
