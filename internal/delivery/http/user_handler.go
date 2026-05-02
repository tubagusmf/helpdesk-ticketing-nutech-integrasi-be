package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
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
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
	group.PUT("/online-status", handler.UpdateOnlineStatus, AuthMiddleware)
	group.GET("/me", handler.GetMe, AuthMiddleware)
	group.PUT("/force-offline/:id", handler.ForceOffline, AuthMiddleware)
	group.PUT("/heartbeat", handler.Heartbeat, AuthMiddleware)
	group.PUT("/logout", handler.Logout)
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
	var filter model.User

	filter.Name = c.QueryParam("name")

	if email := c.QueryParam("email"); email != "" {
		filter.Email = email
	}

	if isActive := c.QueryParam("is_active"); isActive == "true" {
		filter.IsActive = true
	}

	if roleID := c.QueryParam("role_id"); roleID != "" {
		id, _ := strconv.ParseInt(roleID, 10, 64)
		filter.RoleID = id
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	limit := 10

	users, total, err := h.userUsecase.FindAll(
		c.Request().Context(),
		filter,
		page,
		limit,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "users fetched successfully",
		"data":       users,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
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

func (h *UserHandler) UpdateOnlineStatus(c echo.Context) error {
	var body struct {
		IsOnline bool `json:"is_online"`
	}

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim := claimValue.(*model.CustomClaims)

	err := h.userUsecase.UpdateOnlineStatus(
		c.Request().Context(),
		claim.UserID,
		body.IsOnline,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "online status updated",
	})
}

func (h *UserHandler) GetMe(c echo.Context) error {
	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim := claimValue.(*model.CustomClaims)

	user, err := h.userUsecase.FindByID(c.Request().Context(), claim.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": user,
	})
}

func (h *UserHandler) ForceOffline(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	err := h.userUsecase.UpdateOnlineStatus(
		c.Request().Context(),
		id,
		false,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "force offline success",
	})
}

func (h *UserHandler) Heartbeat(c echo.Context) error {
	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim := claimValue.(*model.CustomClaims)

	err := h.userUsecase.UpdateLastSeen(
		c.Request().Context(),
		claim.UserID,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *UserHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	var body struct {
		Token string `json:"token"`
	}

	_ = c.Bind(&body)

	if body.Token != "" {
		var claim model.CustomClaims
		_ = helper.DecodeToken(body.Token, &claim)

		if claim.UserID != 0 {
			_ = h.userUsecase.UpdateOnlineStatus(ctx, claim.UserID, false)
		}
	}

	return c.NoContent(http.StatusOK)
}
