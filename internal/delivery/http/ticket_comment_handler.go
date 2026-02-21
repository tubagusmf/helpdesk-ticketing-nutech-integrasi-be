package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketCommentHandler struct {
	usecase model.ITicketCommentUsecase
}

func NewTicketCommentHandler(e *echo.Echo, u model.ITicketCommentUsecase) {
	handler := &TicketCommentHandler{
		usecase: u,
	}

	group := e.Group("/v1/tickets")

	group.POST("/:id/comments", handler.Create, AuthMiddleware)
	group.GET("/:id/comments", handler.GetByTicketID, AuthMiddleware)
}

func (h *TicketCommentHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	idParam := c.Param("id")
	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	claims, ok := ctx.Value(model.BearerAuthKey).(*model.CustomClaims)
	if !ok || claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}

	comment := model.TicketComment{
		TicketID: ticketID,
		UserID:   claims.UserID,
		Message:  req.Message,
	}

	result, err := h.usecase.Create(ctx, comment)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *TicketCommentHandler) GetByTicketID(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	comments, err := h.usecase.FindByTicketID(
		c.Request().Context(),
		ticketID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, comments)
}
