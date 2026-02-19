package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type LocationHandler struct {
	locationUsecase model.ILocationUsecase
}

func NewLocationHandler(e *echo.Echo, locationUsecase model.ILocationUsecase) {
	handler := &LocationHandler{
		locationUsecase: locationUsecase,
	}

	group := e.Group("/v1/locations")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *LocationHandler) Create(c echo.Context) error {
	var body model.LocationInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	location, err := h.locationUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "location created successfully",
		"data":    location,
	})
}

func (h *LocationHandler) FindAll(c echo.Context) error {
	var filter model.Location

	// Optional query params
	filter.Name = c.QueryParam("name")

	if projectID := c.QueryParam("project_id"); projectID != "" {
		id, err := strconv.ParseInt(projectID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid project_id")
		}
		filter.ProjectID = id
	}

	locations, err := h.locationUsecase.FindAll(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "locations fetched successfully",
		"data":    locations,
	})
}

func (h *LocationHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	location, err := h.locationUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "location fetched successfully",
		"data":    location,
	})
}

func (h *LocationHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateLocationInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.locationUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "location updated successfully",
	})
}

func (h *LocationHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.locationUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "location deleted successfully",
	})
}
