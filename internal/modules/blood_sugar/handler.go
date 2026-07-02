package blood_sugar

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type BloodSugarHandler struct {
	svc BloodSugarService
	log *zap.Logger
}

func NewBloodSugarHandler(svc BloodSugarService, log *zap.Logger) *BloodSugarHandler {
	return &BloodSugarHandler{svc: svc, log: log}
}

// Log handles POST /api/v1/patient/blood-sugar
// @Summary      Log blood sugar measurement
// @Tags         blood-sugar
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  LogBloodSugarRequest  true  "Logging payload"
// @Success      201  {object}  map[string]any
// @Router       /patient/blood-sugar [post]
func (h *BloodSugarHandler) Log(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req LogBloodSugarRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.LogBloodSugar(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "blood sugar logged successfully", res)
}

// GetHistory handles GET /api/v1/patient/blood-sugar
// @Summary      Get own blood sugar history
// @Tags         blood-sugar
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number"
// @Param        limit  query  int  false  "Limit"
// @Success      200  {object}  map[string]any
// @Router       /patient/blood-sugar [get]
func (h *BloodSugarHandler) GetHistory(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 10
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	items, total, err := h.svc.GetPatientHistory(c.Context(), claims.UserID, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "blood sugar logs retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetPatientHistory handles GET /api/v1/admin/patients/:id/blood-sugar
// @Summary      Get patient blood sugar history (Admin/Puskesmas)
// @Tags         blood-sugar
// @Security     BearerAuth
// @Produce      json
// @Param        id     path   string  true   "Patient ID"
// @Param        page   query  int     false  "Page number"
// @Param        limit  query  int     false  "Limit"
// @Success      200  {object}  map[string]any
// @Router       /admin/patients/{id}/blood-sugar [get]
func (h *BloodSugarHandler) GetPatientHistory(c fiber.Ctx) error {
	patientID := c.Params("id")
	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 10
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	items, total, err := h.svc.GetPatientHistory(c.Context(), patientID, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "blood sugar logs retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetDashboard handles GET /api/v1/puskesmas/dashboard/blood-sugar
// @Summary      Get puskesmas monitoring stats
// @Tags         blood-sugar
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /puskesmas/dashboard/blood-sugar [get]
func (h *BloodSugarHandler) GetDashboard(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetPuskesmasDashboard(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "puskesmas dashboard statistics retrieved", res)
}
