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
		splitAuth := strings.Split(authHeader, " ")
		if len(splitAuth) != 2 || splitAuth[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		accessToken := splitAuth[1]

		var claim model.CustomClaims
		if err := helper.DecodeToken(accessToken, &claim); err != nil {
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
