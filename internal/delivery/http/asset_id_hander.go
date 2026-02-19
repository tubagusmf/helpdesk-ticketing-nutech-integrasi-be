package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type AssetIDHandler struct {
	assetUsecase model.IAssetIDUsecase
}

func NewAssetIDHandler(e *echo.Echo, assetUsecase model.IAssetIDUsecase) {
	handler := &AssetIDHandler{
		assetUsecase: assetUsecase,
	}

	group := e.Group("/v1/asset-id")

	group.POST("/create", handler.Create, AuthMiddleware)
	group.GET("", handler.FindAll, AuthMiddleware)
	group.GET("/:id", handler.FindByID, AuthMiddleware)
	group.PUT("/update/:id", handler.Update, AuthMiddleware)
	group.DELETE("/delete/:id", handler.Delete, AuthMiddleware)
}

func (h *AssetIDHandler) Create(c echo.Context) error {
	var body model.AssetIDInput

	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	asset, err := h.assetUsecase.Create(c.Request().Context(), body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "asset_id created successfully",
		"data":    asset,
	})
}

func (h *AssetIDHandler) FindAll(c echo.Context) error {
	var filter model.AssetID

	filter.Name = c.QueryParam("name")

	if partID := c.QueryParam("part_id"); partID != "" {
		id, err := strconv.ParseInt(partID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid part_id")
		}
		filter.PartID = id
	}

	assets, err := h.assetUsecase.FindAll(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "asset_id fetched successfully",
		"data":    assets,
	})
}

func (h *AssetIDHandler) FindByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	asset, err := h.assetUsecase.FindByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "asset_id fetched successfully",
		"data":    asset,
	})
}

func (h *AssetIDHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var body model.UpdateAssetIDInput
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.assetUsecase.Update(c.Request().Context(), id, body); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "asset_id updated successfully",
	})
}

func (h *AssetIDHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.assetUsecase.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "asset_id deleted successfully",
	})
}
