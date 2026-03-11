package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"

	"github.com/labstack/echo/v4"
)

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
			if strings.Contains(err.Error(), "expired") {
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
