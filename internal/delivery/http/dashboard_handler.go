package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/usecase"
)

type DashboardHandler struct {
	usecase *usecase.DashboardUsecase
}

func NewDashboardHandler(e *echo.Echo, u *usecase.DashboardUsecase) {
	handler := &DashboardHandler{usecase: u}

	group := e.Group("/v1/dashboard", AuthMiddleware)

	group.GET("/summary", handler.GetSummary)
	group.GET("/status-distribution", handler.GetStatus)
	group.GET("/priority", handler.GetPriority)
	group.GET("/volume-project", handler.GetVolume)
}

func (h *DashboardHandler) buildFilter(c echo.Context) map[string]interface{} {
	claims := helper.GetUserFromContext(c.Request().Context())

	filter := map[string]interface{}{
		"project_id": c.QueryParam("project_id"),
		"part_id":    c.QueryParam("part_id"),
		"start_date": c.QueryParam("start_date"),
		"end_date":   c.QueryParam("end_date"),
	}

	if claims != nil {
		role := strings.ToLower(claims.Role)

		if role != "administrator" {
			filter["user_id"] = claims.UserID
			filter["role"] = role
		}
	}

	return filter
}

func (h *DashboardHandler) GetSummary(c echo.Context) error {
	filter := h.buildFilter(c)

	data, err := h.usecase.GetSummary(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetStatus(c echo.Context) error {
	filter := h.buildFilter(c)

	data, err := h.usecase.GetStatus(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetPriority(c echo.Context) error {
	filter := h.buildFilter(c)

	data, err := h.usecase.GetPriority(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetVolume(c echo.Context) error {
	filter := h.buildFilter(c)

	data, err := h.usecase.GetVolume(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, data)
}
