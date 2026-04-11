package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
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
	claim, ok := c.Request().Context().
		Value(model.BearerAuthKey).(*model.CustomClaims)

	if !ok || claim == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	projectID, _ := strconv.ParseInt(c.FormValue("project_id"), 10, 64)
	locationID, _ := strconv.ParseInt(c.FormValue("location_id"), 10, 64)
	partID, _ := strconv.ParseInt(c.FormValue("part_id"), 10, 64)
	assetID, _ := strconv.ParseInt(c.FormValue("asset_id"), 10, 64)
	assignedID, _ := strconv.ParseInt(c.FormValue("assigned_to_id"), 10, 64)

	priority := model.TicketPriority(c.FormValue("priority"))
	description := c.FormValue("description")

	input := model.CreateTicketInput{
		ProjectID:    projectID,
		LocationID:   locationID,
		PartID:       partID,
		AssetID:      assetID,
		AssignedToID: assignedID,
		Priority:     priority,
		Description:  description,
	}

	var attachmentURL *string

	fileHeader, err := c.FormFile("attachment")
	if err == nil {

		if fileHeader.Size > 2*1024*1024 {
			return echo.NewHTTPError(http.StatusBadRequest, "file max 2MB")
		}

		file, err := fileHeader.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
		}
		defer file.Close()

		folder := fmt.Sprintf("tickets/project_%d", projectID)
		url, err := helper.UploadImage(file, folder)
		if err != nil {
			log.Println("Failed to upload image:", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "upload failed")
		}

		attachmentURL = &url
	}

	ticket, err := h.ticketUsecase.Create(
		c.Request().Context(),
		claim.UserID,
		input,
		attachmentURL,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "ticket created successfully",
		"data":    ticket,
	})
}

func (h *TicketHandler) FindAll(c echo.Context) error {
	projectID, _ := strconv.ParseInt(c.QueryParam("project_id"), 10, 64)
	staffID, _ := strconv.ParseInt(c.QueryParam("assigned_to_id"), 10, 64)
	reporterID, _ := strconv.ParseInt(c.QueryParam("reporter_id"), 10, 64)

	priority := c.QueryParam("priority")
	status := c.QueryParam("status")
	search := c.QueryParam("search")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}
	limit := 10

	filter := model.Ticket{
		ProjectID:    projectID,
		AssignedToID: staffID,
		ReporterID:   reporterID,
		Priority:     model.TicketPriority(priority),
		Status:       model.TicketStatus(status),
	}

	tickets, total, err := h.ticketUsecase.FindAll(
		c.Request().Context(),
		filter,
		search,
		page,
		limit,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":       tickets,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
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

	fmt.Println("HANDLER ATTACHMENT:", ticket.Attachment)

	return c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateStatus(c echo.Context) error {
	idParam := c.Param("id")

	ticketID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ticket id")
	}

	var req model.UpdateTicketStatusInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	claimValue := c.Request().Context().Value(model.BearerAuthKey)
	if claimValue == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	claim := claimValue.(*model.CustomClaims)
	userID := claim.UserID

	err = h.ticketUsecase.UpdateStatus(
		c.Request().Context(),
		ticketID,
		userID,
		req,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "status updated successfully",
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
