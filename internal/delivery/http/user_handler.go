package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type UserHandler struct {
	userUsecase model.IUserUsecase
}

func NewUserHandler(e *echo.Echo, userUsecase model.IUserUsecase) {
	handler := &UserHandler{
		userUsecase: userUsecase,
	}

	group := e.Group("/v1/users")

	group.POST("/login", handler.Login)
	group.POST("/register", handler.Create)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/:id", handler.Update, AuthMiddleware)
	group.DELETE("/:id", handler.Delete, AuthMiddleware)
}

func (h *UserHandler) Login(c echo.Context) error {
	var body model.LoginInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.userUsecase.Login(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "login success",
		"token":   token,
	})
}

func (h *UserHandler) Create(c echo.Context) error {
	var body model.CreateUserInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.userUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "register success",
		"token":   token,
	})
}

func (h *UserHandler) FindAll(c echo.Context) error {
	users, err := h.userUsecase.FindAll(c.Request().Context(), model.User{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}

func (h *UserHandler) FindByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	user, err := h.userUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var body model.UpdateUserInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.userUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "updated successfully",
	})
}

func (h *UserHandler) Delete(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.userUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "deleted successfully",
	})
}
