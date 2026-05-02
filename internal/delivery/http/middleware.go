package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"

	"github.com/labstack/echo/v4"
)

var userUC model.IUserUsecase

func InitAuthMiddleware(uc model.IUserUsecase) {
	userUC = uc
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
		}

		splitAuth := strings.Split(authHeader, " ")
		if len(splitAuth) != 2 || splitAuth[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
		}

		accessToken := splitAuth[1]

		var claim model.CustomClaims
		err := helper.DecodeToken(accessToken, &claim)

		if err != nil {

			if errors.Is(err, jwt.ErrTokenExpired) {

				if claim.UserID != 0 && userUC != nil {
					_ = userUC.UpdateOnlineStatus(
						context.Background(),
						claim.UserID,
						false,
					)
				}

				return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		ctx := context.WithValue(
			c.Request().Context(),
			model.BearerAuthKey,
			&claim,
		)

		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
