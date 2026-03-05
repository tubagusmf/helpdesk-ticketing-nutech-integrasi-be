package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type SolutionHandler struct {
	solutionUsecase model.ISolutionUsecase
}

func NewSolutionHandler(e *echo.Echo, solutionUsecase model.ISolutionUsecase) {
	handler := &SolutionHandler{
		solutionUsecase: solutionUsecase,
	}

	group := e.Group("/v1/solutions")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *SolutionHandler) Create(c echo.Context) error {
	var body model.CreateSolutionInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	solution, err := h.solutionUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "solution created successfully",
		"data":    solution,
	})
}

func (h *SolutionHandler) FindAll(c echo.Context) error {
	var filter model.Solution

	filter.Name = c.QueryParam("name")

	if causeID := c.QueryParam("cause_id"); causeID != "" {
		id, err := strconv.ParseInt(causeID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid cause_id")
		}
		filter.CauseID = id
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	limit := 10

	solutions, total, err := h.solutionUsecase.FindAll(c.Request().Context(), filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "solutions fetched successfully",
		"data":       solutions,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
}

func (h *SolutionHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	solution, err := h.solutionUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, solution)
}

func (h *SolutionHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateSolutionInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.solutionUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "solution updated successfully",
	})
}

func (h *SolutionHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.solutionUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "solution deleted successfully",
	})
}
