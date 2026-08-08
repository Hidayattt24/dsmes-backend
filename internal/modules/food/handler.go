package food

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type FoodHandler struct {
	svc FoodService
	log *zap.Logger
}

func NewFoodHandler(svc FoodService, log *zap.Logger) *FoodHandler {
	return &FoodHandler{svc: svc, log: log}
}

// GetFoods handles GET /api/v1/admin/foods and GET /api/v1/foods
func (h *FoodHandler) GetFoods(c fiber.Ctx) error {
	var query FoodFilterQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.NewBadRequest("invalid query parameters")
	}

	items, total, err := h.svc.GetFoods(c.Context(), query)
	if err != nil {
		return err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages <= 0 {
		totalPages = 1
	}

	return response.SuccessWithMeta(c, "food list retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// SearchFoods handles GET /api/v1/foods/search
func (h *FoodHandler) SearchFoods(c fiber.Ctx) error {
	var query FoodFilterQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.NewBadRequest("invalid query parameters")
	}

	items, total, err := h.svc.SearchFoods(c.Context(), query)
	if err != nil {
		return err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages <= 0 {
		totalPages = 1
	}

	return response.SuccessWithMeta(c, "food search results retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/foods/:id and GET /api/v1/admin/foods/:id
func (h *FoodHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "stats" {
		return h.GetStats(c)
	}
	if id == "export" {
		return h.Export(c)
	}
	item, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "food item retrieved", item)
}

// Create handles POST /api/v1/admin/foods
func (h *FoodHandler) Create(c fiber.Ctx) error {
	var req CreateFoodRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "food item created", res)
}

// Update handles PUT /api/v1/admin/foods/:id
func (h *FoodHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateFoodRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "food item updated", res)
}

// Delete handles DELETE /api/v1/admin/foods/:id
func (h *FoodHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}

// PreviewImport handles POST /api/v1/admin/foods/import/preview
func (h *FoodHandler) PreviewImport(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return errs.NewBadRequest("file is required (form-data field name 'file')")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return errs.NewBadRequest("failed to open uploaded file")
	}
	defer file.Close()

	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		return errs.NewBadRequest("failed to read file content")
	}

	res, err := h.svc.PreviewExcelImport(c.Context(), fileBytes)
	if err != nil {
		return errs.NewUnprocessable(err.Error())
	}

	return response.Success(c, "excel preview generated", res)
}

// ConfirmImport handles POST /api/v1/admin/foods/import/confirm
func (h *FoodHandler) ConfirmImport(c fiber.Ctx) error {
	var req ExcelImportConfirmRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.ConfirmExcelImport(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, fmt.Sprintf("imported %d foods successfully", res.SuccessCount), res)
}

// ExportHandles handles GET /api/v1/admin/foods/export
func (h *FoodHandler) Export(c fiber.Ctx) error {
	var query FoodFilterQuery
	if err := c.Bind().Query(&query); err != nil {
		return errs.NewBadRequest("invalid query parameters")
	}
	format := c.Query("format", "xlsx")

	fileBytes, contentType, fileName, err := h.svc.ExportFoods(c.Context(), query, format)
	if err != nil {
		return err
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	return c.Send(fileBytes)
}

// GetStats handles GET /api/v1/admin/foods/stats
func (h *FoodHandler) GetStats(c fiber.Ctx) error {
	stats, err := h.svc.GetStats(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "food statistics retrieved", stats)
}
