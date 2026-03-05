package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type CauseHandler struct {
	causeUsecase model.ICauseUsecase
}

func NewCauseHandler(e *echo.Echo, causeUsecase model.ICauseUsecase) {
	handler := &CauseHandler{
		causeUsecase: causeUsecase,
	}

	group := e.Group("/v1/causes")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *CauseHandler) Create(c echo.Context) error {
	var body model.CreateCauseInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	cause, err := h.causeUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "cause created successfully",
		"data":    cause,
	})
}

func (h *CauseHandler) FindAll(c echo.Context) error {
	var filter model.Cause

	filter.Name = c.QueryParam("name")

	if partID := c.QueryParam("part_id"); partID != "" {
		id, err := strconv.ParseInt(partID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid part_id")
		}
		filter.PartID = id
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	limit := 10

	causes, total, err := h.causeUsecase.FindAll(c.Request().Context(), filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "causes fetched successfully",
		"data":       causes,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
}

func (h *CauseHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	cause, err := h.causeUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, cause)
}

func (h *CauseHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateCauseInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.causeUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "cause updated successfully",
	})
}

func (h *CauseHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.causeUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "cause deleted successfully",
	})
}
