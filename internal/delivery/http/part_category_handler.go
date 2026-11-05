package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type PartCategoryHandler struct {
	categoryUsecase model.IPartCategoryUsecase
}

func NewPartCategoryHandler(e *echo.Echo, categoryUsecase model.IPartCategoryUsecase) {
	handler := &PartCategoryHandler{
		categoryUsecase: categoryUsecase,
	}

	group := e.Group("/v1/part-categories")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/part/:part_id", handler.FindByPartID, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *PartCategoryHandler) Create(c echo.Context) error {
	var body model.PartCategoryInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	category, err := h.categoryUsecase.Create(
		c.Request().Context(),
		body,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "part category created successfully",
		"data":    category,
	})
}

func (h *PartCategoryHandler) FindAll(c echo.Context) error {
	var filter model.PartCategory

	filter.Name = c.QueryParam("name")

	if partID := c.QueryParam("part_id"); partID != "" {
		id, err := strconv.ParseInt(partID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"invalid part_id",
			)
		}

		filter.PartID = id
	}

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil && c.QueryParam("page") != "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid page")
	}

	if page < 1 {
		page = 1
	}

	limit := 10

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"invalid limit",
			)
		}

		if parsedLimit > 1000 {
			parsedLimit = 1000
		}

		limit = parsedLimit
	}

	categories, total, err := h.categoryUsecase.FindAll(
		c.Request().Context(),
		filter,
		page,
		limit,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	totalPage := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "part categories fetched successfully",
		"data":       categories,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
}

func (h *PartCategoryHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	category, err := h.categoryUsecase.FindByID(
		c.Request().Context(),
		id,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "part category fetched successfully",
		"data":    category,
	})
}

func (h *PartCategoryHandler) FindByPartID(c echo.Context) error {
	partID, err := strconv.ParseInt(c.Param("part_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid part_id",
		)
	}

	categories, err := h.categoryUsecase.FindByPartID(
		c.Request().Context(),
		partID,
	)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "part categories fetched successfully",
		"data":    categories,
	})
}

func (h *PartCategoryHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdatePartCategoryInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.categoryUsecase.Update(
		c.Request().Context(),
		id,
		body,
	); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "part category updated successfully",
	})
}

func (h *PartCategoryHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.categoryUsecase.Delete(
		c.Request().Context(),
		id,
	); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "part category deleted successfully",
	})
}
