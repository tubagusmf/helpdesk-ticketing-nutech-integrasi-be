package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketResolutionHandler struct {
	usecase model.ITicketResolutionUsecase
}

func NewTicketResolutionHandler(e *echo.Echo, u model.ITicketResolutionUsecase) {
	handler := &TicketResolutionHandler{
		usecase: u,
	}

	group := e.Group("/v1/tickets", AuthMiddleware)

	group.POST("/:id/resolution", handler.Create)
	group.GET("/:id/resolution", handler.GetByTicketID)
}

func (h *TicketResolutionHandler) Create(c echo.Context) error {
	idParam := c.Param("id")
	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	var req model.CreateTicketResolutionInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.TicketID = ticketID

	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found in token")
	}

	claim := claimValue.(*model.CustomClaims)
	userID := claim.UserID

	resolution, err := h.usecase.Create(c.Request().Context(), userID, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, resolution)
}

func (h *TicketResolutionHandler) GetByTicketID(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	resolution, err := h.usecase.FindByTicketID(
		c.Request().Context(),
		ticketID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, resolution)
}
