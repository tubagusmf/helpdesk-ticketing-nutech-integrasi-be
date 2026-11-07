package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type ProblemHandler struct {
	problemUsecase model.IProblemUsecase
}

func NewProblemHandler(e *echo.Echo, problemUsecase model.IProblemUsecase) {
	handler := &ProblemHandler{
		problemUsecase: problemUsecase,
	}

	group := e.Group("/v1/problems")

	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/part/:part_id", handler.FindByPartID, AuthMiddleware)
	group.POST("/create", handler.Create, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *ProblemHandler) checkAdmin(c echo.Context) error {
	claimValue := c.Request().
		Context().
		Value(model.BearerAuthKey)

	claim, ok := claimValue.(*model.CustomClaims)

	if !ok || claim == nil {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"unauthorized",
		)
	}

	if claim.Role != "ADMINISTRATOR" {
		return echo.NewHTTPError(
			http.StatusForbidden,
			"akses hanya untuk Administrator",
		)
	}

	return nil
}

func (h *ProblemHandler) FindAll(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit != 10 && limit != 25 && limit != 50 {
		limit = 10
	}

	search := strings.TrimSpace(c.QueryParam("name"))

	data, total, err := h.problemUsecase.FindAll(
		c.Request().Context(),
		search,
		page,
		limit,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err.Error(),
		)
	}

	totalPage := int(
		(total + int64(limit) - 1) / int64(limit),
	)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":       data,
		"page":       page,
		"total_data": total,
		"total_page": totalPage,
	})
}

func (h *ProblemHandler) FindByPartID(c echo.Context) error {
	partID, err := strconv.ParseInt(
		c.Param("part_id"),
		10,
		64,
	)

	if err != nil || partID <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Part ID tidak valid",
		)
	}

	data, err := h.problemUsecase.FindByPartID(
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
		"message": "data permasalahan berhasil diambil",
		"data":    data,
	})
}

func (h *ProblemHandler) Create(c echo.Context) error {
	if err := h.checkAdmin(c); err != nil {
		return err
	}

	var input model.ProblemInput

	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"request tidak valid",
		)
	}

	problem, err := h.problemUsecase.Create(
		c.Request().Context(),
		input,
	)

	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			err.Error(),
		)
	}

	return c.JSON(
		http.StatusCreated,
		map[string]interface{}{
			"message": "permasalahan berhasil ditambahkan",
			"data":    problem,
		},
	)
}

func (h *ProblemHandler) Update(c echo.Context) error {
	if err := h.checkAdmin(c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"ID permasalahan tidak valid",
		)
	}

	var input model.ProblemInput

	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"request tidak valid",
		)
	}

	problem, err := h.problemUsecase.Update(
		c.Request().Context(),
		id,
		input,
	)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		return echo.NewHTTPError(status, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "permasalahan berhasil diperbarui",
		"data":    problem,
	})
}

func (h *ProblemHandler) Delete(c echo.Context) error {
	if err := h.checkAdmin(c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil || id <= 0 {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"ID permasalahan tidak valid",
		)
	}

	err = h.problemUsecase.Delete(c.Request().Context(), id)

	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}

		return echo.NewHTTPError(status, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "permasalahan berhasil dihapus",
	})
}
