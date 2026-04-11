package http

import (
	"fmt"
	"net/http"
	"strconv"
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

	causeID, _ := strconv.ParseInt(c.FormValue("cause_id"), 10, 64)
	solutionID, _ := strconv.ParseInt(c.FormValue("solution_id"), 10, 64)
	notes := c.FormValue("resolution_notes")
	completionTimeStr := c.FormValue("completion_time")

	var completionTime time.Time
	if completionTimeStr != "" {
		completionTime, _ = time.Parse("2006-01-02T15:04", completionTimeStr)
	}

	fileHeader, err := c.FormFile("attachment")
	var attachmentURL string

	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed open file")
		}
		defer file.Close()

		folder := fmt.Sprintf("tickets/resolution_%d", ticketID)

		url, err := helper.UploadImage(file, folder)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "upload failed")
		}

		attachmentURL = url
	}

	status := model.TicketStatus(c.FormValue("status"))

	req := model.CreateTicketResolutionInput{
		TicketID:        ticketID,
		CauseID:         causeID,
		SolutionID:      solutionID,
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
