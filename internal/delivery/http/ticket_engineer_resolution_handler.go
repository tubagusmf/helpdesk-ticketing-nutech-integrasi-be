package http

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/usecase"
)

type TicketEngineerResolutionHandler struct {
	usecase *usecase.TicketEngineerResolutionUsecase
}

func NewTicketEngineerResolutionHandler(e *echo.Echo, u *usecase.TicketEngineerResolutionUsecase) {
	handler := &TicketEngineerResolutionHandler{
		usecase: u,
	}

	group := e.Group("/v1/tickets", AuthMiddleware)

	group.POST("/:id/engineer-resolution", handler.Create)
	group.GET("/:id/engineer-resolution", handler.GetByTicketID)
}

func (h *TicketEngineerResolutionHandler) Create(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	claimValue := c.Request().Context().Value(
		model.BearerAuthKey,
	)

	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim, ok := claimValue.(*model.CustomClaims)
	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication claims")
	}

	engineerID := claim.UserID

	solution := strings.TrimSpace(
		c.FormValue("solution"),
	)

	if solution == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "solution is required")
	}

	var files []*multipart.FileHeader

	form, err := c.MultipartForm()

	if err == nil && form != nil {
		files = form.File["attachments"]
	}

	resolution, err := h.usecase.SubmitResolution(c.Request().Context(), ticketID, engineerID, solution, files)

	if err != nil {
		errMessage := err.Error()

		switch {
		case strings.Contains(
			errMessage,
			"solution wajib diisi",
		):
			return echo.NewHTTPError(http.StatusBadRequest, errMessage)

		case strings.Contains(
			errMessage,
			"melebihi ukuran maksimal",
		):
			return echo.NewHTTPError(http.StatusBadRequest, errMessage)

		case strings.Contains(
			errMessage,
			"hanya engineer",
		):
			return echo.NewHTTPError(http.StatusForbidden, errMessage)

		case strings.Contains(
			errMessage,
			"bukan merupakan ticket",
		):
			return echo.NewHTTPError(http.StatusForbidden, errMessage)

		case strings.Contains(
			errMessage,
			"engineer tidak ditemukan",
		):
			return echo.NewHTTPError(http.StatusNotFound, errMessage)

		case strings.Contains(errMessage, "ticket tidak ditemukan"):
			return echo.NewHTTPError(http.StatusNotFound, errMessage)

		default:
			return echo.NewHTTPError(http.StatusInternalServerError, errMessage)
		}
	}

	return c.JSON(http.StatusCreated, resolution)
}

func (h *TicketEngineerResolutionHandler) GetByTicketID(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	resolution, err := h.usecase.GetByTicketID(c.Request().Context(), ticketID)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, resolution)
}
