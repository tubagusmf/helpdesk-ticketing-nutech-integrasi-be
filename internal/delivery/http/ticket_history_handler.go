package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketHistoryHandler struct {
	usecase model.ITicketHistoryUsecase
}

func NewTicketHistoryHandler(e *echo.Echo, u model.ITicketHistoryUsecase) {
	handler := &TicketHistoryHandler{
		usecase: u,
	}

	group := e.Group("/v1/tickets")

	group.GET("/history/:id", handler.GetByTicketID, AuthMiddleware)
}

func (h *TicketHistoryHandler) GetByTicketID(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	histories, err := h.usecase.FindByTicketID(
		c.Request().Context(),
		ticketID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, histories)
}
