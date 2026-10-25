package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
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

	group.GET("/:id/history", handler.GetByTicketID, AuthMiddleware)
	group.GET("/:id/history/export", handler.ExportByTicketID, AuthMiddleware)
	group.GET("/:id/history/export-pdf", handler.ExportPDFByTicketID, AuthMiddleware)
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": histories,
	})
}

func (h *TicketHistoryHandler) ExportByTicketID(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid ticket id",
		)
	}

	histories, err := h.usecase.FindByTicketID(
		c.Request().Context(),
		ticketID,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal mengambil history tiket",
		)
	}

	ticketCode := strings.TrimSpace(c.QueryParam("ticket_code"))
	if ticketCode == "" {
		ticketCode = fmt.Sprintf("TICKET-%d", ticketID)
	}

	buffer, err := helper.GenerateExcelTicketHistory(
		ticketCode,
		ticketID,
		histories,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal membuat laporan Excel history tiket",
		)
	}

	fileName := fmt.Sprintf("ticket_history_%s.xlsx", ticketCode)

	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, fileName),
	)
	c.Response().Header().Set("Cache-Control", "no-store")

	return c.Blob(
		http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		buffer.Bytes(),
	)
}

func (h *TicketHistoryHandler) ExportPDFByTicketID(c echo.Context) error {
	ticketID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || ticketID <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid ticket id",
		)
	}

	histories, err := h.usecase.FindByTicketID(
		c.Request().Context(),
		ticketID,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal mengambil history tiket",
		)
	}

	ticketCode := strings.TrimSpace(c.QueryParam("ticket_code"))
	if ticketCode == "" {
		ticketCode = fmt.Sprintf("TICKET-%d", ticketID)
	}

	buffer, err := helper.GeneratePDFTicketHistory(
		ticketCode,
		ticketID,
		histories,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"gagal membuat laporan PDF history tiket",
		)
	}

	// Hindari karakter yang berisiko merusak nama file.
	safeTicketCode := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		"\"", "",
		"\r", "",
		"\n", "",
	).Replace(ticketCode)

	fileName := fmt.Sprintf(
		"ticket_history_%s.pdf",
		safeTicketCode,
	)

	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, fileName),
	)
	c.Response().Header().Set("Cache-Control", "no-store")

	return c.Blob(
		http.StatusOK,
		"application/pdf",
		buffer.Bytes(),
	)
}
