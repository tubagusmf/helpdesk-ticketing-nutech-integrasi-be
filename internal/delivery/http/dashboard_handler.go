package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/usecase"
)

type DashboardHandler struct {
	usecase *usecase.DashboardUsecase
}

func NewDashboardHandler(e *echo.Echo, u *usecase.DashboardUsecase) {
	handler := &DashboardHandler{
		usecase: u,
	}

	group := e.Group(
		"/v1/dashboard",
		AuthMiddleware,
	)

	group.GET("/summary", handler.GetSummary)
	group.GET("/status-distribution", handler.GetStatus)
	group.GET("/priority", handler.GetPriority)
	group.GET("/volume-project", handler.GetVolume)
}

func (h *DashboardHandler) buildFilter(c echo.Context) (model.DashboardFilter, error) {
	projectID, err := parseInt64Query(c, "project_id")
	if err != nil {
		return model.DashboardFilter{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"project_id harus berupa angka",
		)
	}

	partID, err := parseInt64Query(c, "part_id")
	if err != nil {
		return model.DashboardFilter{}, echo.NewHTTPError(
			http.StatusBadRequest,
			"part_id harus berupa angka",
		)
	}

	claims := helper.GetUserFromContext(
		c.Request().Context(),
	)

	filter := model.DashboardFilter{
		ProjectID: projectID,
		PartID:    partID,
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
	}

	if claims != nil {
		role := strings.ToUpper(claims.Role)

		filter.UserID = claims.UserID
		filter.Role = role
	}

	return filter, nil
}

func (h *DashboardHandler) GetSummary(c echo.Context) error {
	filter, err := h.buildFilter(c)
	if err != nil {
		return err
	}

	data, err := h.usecase.GetSummary(
		c.Request().Context(),
		filter,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetStatus(c echo.Context) error {
	filter, err := h.buildFilter(c)
	if err != nil {
		return err
	}

	data, err := h.usecase.GetStatus(
		c.Request().Context(),
		filter,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetPriority(c echo.Context) error {
	filter, err := h.buildFilter(c)
	if err != nil {
		return err
	}

	data, err := h.usecase.GetPriority(
		c.Request().Context(),
		filter,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *DashboardHandler) GetVolume(c echo.Context) error {
	filter, err := h.buildFilter(c)
	if err != nil {
		return err
	}

	data, err := h.usecase.GetVolume(
		c.Request().Context(),
		filter,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, data)
}

func parseInt64Query(c echo.Context, key string) (int64, error) {
	value := c.QueryParam(key)

	if value == "" {
		return 0, nil
	}

	return strconv.ParseInt(value, 10, 64)
}
