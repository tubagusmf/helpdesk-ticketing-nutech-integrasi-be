package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type RoleHandler struct {
	roleUsecase model.IRoleUsecase
}

func NewRoleHandler(e *echo.Echo, roleUsecase model.IRoleUsecase) {
	handler := &RoleHandler{
		roleUsecase: roleUsecase,
	}

	group := e.Group("/v1/roles")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *RoleHandler) Create(c echo.Context) error {
	var body model.CreateRoleInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	role, err := h.roleUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "role created successfully",
		"data":    role,
	})
}

func (h *RoleHandler) FindAll(c echo.Context) error {
	roles, err := h.roleUsecase.FindAll(c.Request().Context(), model.Role{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, roles)
}

func (h *RoleHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	role, err := h.roleUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateRoleInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.roleUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "role updated successfully",
	})
}

func (h *RoleHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.roleUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "role deleted successfully",
	})
}
