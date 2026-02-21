package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketHandler struct {
	ticketUsecase model.ITicketUsecase
}

func NewTicketHandler(e *echo.Echo, ticketUsecase model.ITicketUsecase) {
	handler := &TicketHandler{
		ticketUsecase: ticketUsecase,
	}

	group := e.Group("/v1/tickets")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update-status/:id", handler.UpdateStatus, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *TicketHandler) Create(c echo.Context) error {
	var body model.CreateTicketInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claim, ok := c.Request().Context().Value(model.BearerAuthKey).(*model.CustomClaims)
	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	userID := claim.UserID

	ticket, err := h.ticketUsecase.Create(c.Request().Context(), userID, body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "ticket created successfully",
		"data":    ticket,
	})
}

func (h *TicketHandler) FindAll(c echo.Context) error {
	tickets, err := h.ticketUsecase.FindAll(c.Request().Context(), model.Ticket{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tickets)
}

func (h *TicketHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	ticket, err := h.ticketUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateStatus(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateTicketStatusInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claim, ok := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	userID := claim.UserID

	if err := h.ticketUsecase.UpdateStatus(
		c.Request().Context(),
		id,
		userID,
		body,
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ticket status updated successfully",
	})
}

func (h *TicketHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.ticketUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ticket deleted successfully",
	})
}
