package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type NotificationHandler struct {
	notificationUsecase model.INotificationUsecase
}

func NewNotificationHandler(
	e *echo.Echo,
	notificationUsecase model.INotificationUsecase,
) {
	handler := &NotificationHandler{
		notificationUsecase: notificationUsecase,
	}

	group := e.Group("/v1/notifications")

	group.GET("", handler.FindAllByUserID, AuthMiddleware)
	group.PUT("/:id/read", handler.MarkAsRead, AuthMiddleware)
	group.GET("/unread/count", handler.CountUnread, AuthMiddleware)
}

func (h *NotificationHandler) FindAllByUserID(c echo.Context) error {
	claim := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	notifications, err := h.notificationUsecase.FindAllByUserID(
		c.Request().Context(),
		claim.UserID,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": notifications,
	})
}

func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	claim := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	err = h.notificationUsecase.MarkAsRead(
		c.Request().Context(),
		id,
		claim.UserID,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "notification marked as read",
	})
}

func (h *NotificationHandler) CountUnread(c echo.Context) error {
	claim := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	total, err := h.notificationUsecase.CountUnread(
		c.Request().Context(),
		claim.UserID,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_unread": total,
	})
}
