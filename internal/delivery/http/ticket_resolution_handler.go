package http

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

const maxResolutionAttachmentSize = 10 * 1024 * 1024

type TicketResolutionHandler struct {
	usecase model.ITicketResolutionUsecase
}

func NewTicketResolutionHandler(e *echo.Echo, u model.ITicketResolutionUsecase) {
	handler := &TicketResolutionHandler{usecase: u}

	group := e.Group("/v1/tickets", AuthMiddleware)
	group.POST("/:id/resolution", handler.Create)
	group.GET("/:id/resolution", handler.GetByTicketID)
	group.PUT("/:id/status", handler.UpdateStatus)
}

func (h *TicketResolutionHandler) Create(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	claim, ok := claimValue.(*model.CustomClaims)
	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	var partCategoryID *int64

	rawPartCategoryID := strings.TrimSpace(c.FormValue("part_category_id"))

	if rawPartCategoryID != "" {
		parsed, parseErr := strconv.ParseInt(rawPartCategoryID, 10, 64)

		if parseErr != nil || parsed <= 0 {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"part_category_id tidak valid",
			)
		}

		partCategoryID = &parsed
	}

	var causeID *int64
	rawCauseID := strings.TrimSpace(c.FormValue("cause_id"))
	if rawCauseID != "" {
		parsed, parseErr := strconv.ParseInt(rawCauseID, 10, 64)
		if parseErr != nil || parsed <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "cause_id tidak valid")
		}
		causeID = &parsed
	}

	customCause := strings.TrimSpace(c.FormValue("custom_cause"))
	customSolution := strings.TrimSpace(c.FormValue("custom_solution"))

	var solutionID *int64
	rawSolutionID := strings.TrimSpace(c.FormValue("solution_id"))
	if rawSolutionID != "" {
		parsed, parseErr := strconv.ParseInt(rawSolutionID, 10, 64)
		if parseErr != nil || parsed <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "solution_id tidak valid")
		}
		solutionID = &parsed
	}

	notes := c.FormValue("resolution_notes")
	completionTimeStr := c.FormValue("completion_time")

	var completionTime time.Time
	if completionTimeStr != "" {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "gagal memuat zona waktu Asia/Jakarta")
		}

		completionTime, err = time.ParseInLocation("2006-01-02T15:04", completionTimeStr, loc)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "format waktu penyelesaian tidak valid")
		}
	}

	var attachmentURL string
	fileHeader, err := c.FormFile("attachment")
	if err == nil && fileHeader != nil {
		if fileHeader.Size > maxResolutionAttachmentSize {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "ukuran foto penyelesaian maksimal 10 MB")
		}

		file, openErr := fileHeader.Open()
		if openErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed open file")
		}
		defer file.Close()

		sniff := make([]byte, 512)
		n, readErr := file.Read(sniff)
		if readErr != nil && readErr != io.EOF {
			return echo.NewHTTPError(http.StatusBadRequest, "gagal membaca file gambar")
		}
		contentType := http.DetectContentType(sniff[:n])
		if !strings.HasPrefix(contentType, "image/") {
			return echo.NewHTTPError(http.StatusBadRequest, "lampiran harus berupa file gambar")
		}

		if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "gagal memproses file gambar")
		}

		url, uploadErr := helper.UploadImage(file, "tickets/resolution")
		if uploadErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "upload gagal")
		}
		attachmentURL = url
	} else if err != nil && err != http.ErrMissingFile {
		return echo.NewHTTPError(http.StatusBadRequest, "lampiran tidak valid")
	}

	status := model.TicketStatus(c.FormValue("status"))
	if status != model.StatusResolved {
		return echo.NewHTTPError(http.StatusBadRequest, "status resolusi harus RESOLVED")
	}

	input := model.CreateTicketResolutionInput{
		TicketID:        ticketID,
		PartCategoryID:  partCategoryID,
		CauseID:         causeID,
		CustomCause:     customCause,
		SolutionID:      solutionID,
		CustomSolution:  customSolution,
		ResolutionNotes: notes,
		CompletionTime:  completionTime,
		AttachmentURL:   attachmentURL,
		Status:          status,
	}

	resolution, err := h.usecase.Create(c.Request().Context(), claim.UserID, input)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, resolution)
}

func (h *TicketResolutionHandler) GetByTicketID(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	resolution, err := h.usecase.FindByTicketID(c.Request().Context(), ticketID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, resolution)
}

func (h *TicketResolutionHandler) UpdateStatus(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	var req model.UpdateTicketStatusInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if req.Status == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	claim, ok := c.Request().Context().Value(model.BearerAuthKey).(*model.CustomClaims)
	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	if err := h.usecase.UpdateStatus(c.Request().Context(), ticketID, claim.UserID, req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "status updated")
}
