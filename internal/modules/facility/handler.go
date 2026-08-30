package facility

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type FacilityHandler struct {
	svc FacilityService
	log *zap.Logger
}

func NewFacilityHandler(svc FacilityService, log *zap.Logger) *FacilityHandler {
	return &FacilityHandler{svc: svc, log: log}
}

// List handles GET /api/v1/health-facilities and GET /api/v1/admin/facilities
func (h *FacilityHandler) List(c fiber.Ctx) error {
	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 100
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}
	search := c.Query("search")

	items, total, err := h.svc.ListFacilities(c.Context(), search, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "health facilities retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Create handles POST /api/v1/admin/facilities
func (h *FacilityHandler) Create(c fiber.Ctx) error {
	var req CreateFacilityRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateFacility(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "health facility created", res)
}

// Update handles PUT /api/v1/admin/facilities/:id
func (h *FacilityHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateFacilityRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateFacility(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "health facility updated", res)
}

// Delete handles DELETE /api/v1/admin/facilities/:id
func (h *FacilityHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteFacility(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}
