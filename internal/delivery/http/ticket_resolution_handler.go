package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
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
	group.PUT("/:id/status", handler.UpdateStatus)
}

func (h *TicketResolutionHandler) Create(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim := claimValue.(*model.CustomClaims)
	userID := claim.UserID

	causeID, err := strconv.ParseInt(c.FormValue("cause_id"), 10, 64)
	if err != nil || causeID <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"penyebab wajib dipilih",
		)
	}

	var solutionID *int64

	rawSolutionID := strings.TrimSpace(c.FormValue("solution_id"))

	if rawSolutionID != "" {
		parsedSolutionID, parseErr := strconv.ParseInt(
			rawSolutionID,
			10,
			64,
		)

		if parseErr != nil || parsedSolutionID <= 0 {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"solution_id tidak valid",
			)
		}

		solutionID = &parsedSolutionID
	}

	customSolution := strings.TrimSpace(
		c.FormValue("custom_solution"),
	)

	notes := c.FormValue("resolution_notes")
	completionTimeStr := c.FormValue("completion_time")

	var completionTime time.Time

	if completionTimeStr != "" {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"gagal memuat zona waktu Asia/Jakarta",
			)
		}

		completionTime, err = time.ParseInLocation(
			"2006-01-02T15:04",
			completionTimeStr,
			loc,
		)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"format waktu penyelesaian tidak valid",
			)
		}
	}

	fileHeader, err := c.FormFile("attachment")
	var attachmentURL string

	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed open file")
		}
		defer file.Close()

		folder := "tickets/resolution"

		url, err := helper.UploadImage(file, folder)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				"upload gagal",
			)
		}

		attachmentURL = url
	}

	status := model.TicketStatus(c.FormValue("status"))

	req := model.CreateTicketResolutionInput{
		TicketID:        ticketID,
		CauseID:         causeID,
		SolutionID:      solutionID,
		CustomSolution:  customSolution,
		ResolutionNotes: notes,
		CompletionTime:  completionTime,
		AttachmentURL:   attachmentURL,
		Status:          status,
	}

	resolution, err := h.usecase.Create(
		c.Request().Context(),
		userID,
		req,
	)

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

func (h *TicketResolutionHandler) UpdateStatus(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	var req model.UpdateTicketStatusInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Status == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	claim := c.Request().Context().Value(model.BearerAuthKey).(*model.CustomClaims)
	userID := claim.UserID

	err = h.usecase.UpdateStatus(
		c.Request().Context(),
		ticketID,
		userID,
		req,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "status updated")
}
