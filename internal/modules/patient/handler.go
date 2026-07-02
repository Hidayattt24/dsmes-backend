package patient

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type PatientHandler struct {
	svc PatientService
	log *zap.Logger
}

func NewPatientHandler(svc PatientService, log *zap.Logger) *PatientHandler {
	return &PatientHandler{svc: svc, log: log}
}

// Register handles POST /api/v1/auth/register
// @Summary      Patient registration (signup)
// @Description  Register a new patient and seed default routines & reminders
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  RegisterPatientRequest  true  "Registration payload"
// @Success      201  {object}  map[string]any
// @Router       /auth/register [post]
func (h *PatientHandler) Register(c fiber.Ctx) error {
	var req RegisterPatientRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.RegisterPatient(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "patient registered successfully", res)
}

// List handles GET /api/v1/admin/patients
// @Summary      List all patients (Admin)
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        search  query  string  false  "Search pattern (name/email)"
// @Param        page    query  int     false  "Page number"
// @Param        limit   query  int     false  "Limit"
// @Success      200     {object}  map[string]any
// @Router       /admin/patients [get]
func (h *PatientHandler) List(c fiber.Ctx) error {
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
	search := c.Query("search")

	items, total, err := h.svc.ListPatients(c.Context(), PatientFilterQuery{
		Search: search,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "patients retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ListPuskesmas handles GET /api/v1/puskesmas/patients
// @Summary      List assigned patients (Puskesmas)
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        search  query  string  false  "Search pattern (name/email)"
// @Param        page    query  int     false  "Page number"
// @Param        limit   query  int     false  "Limit"
// @Success      200     {object}  map[string]any
// @Router       /puskesmas/patients [get]
func (h *PatientHandler) ListPuskesmas(c fiber.Ctx) error {
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
	search := c.Query("search")

	items, total, err := h.svc.ListPatients(c.Context(), PatientFilterQuery{
		PuskesmasID: claims.UserID,
		Search:      search,
		Page:        page,
		Limit:       limit,
	})
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "patients retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/admin/patients/:id or /api/v1/puskesmas/patients/:id
// @Summary      Get patient details
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Patient ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/patients/{id} [get]
func (h *PatientHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.GetPatient(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "patient details retrieved", res)
}

// GetMe handles GET /api/v1/patient/me
// @Summary      Get current patient profile
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/me [get]
func (h *PatientHandler) GetMe(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetPatient(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "profile retrieved", res)
}

// UpdateMe handles PUT /api/v1/patient/me
// @Summary      Update current patient profile
// @Tags         patient
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  UpdatePatientProfileRequest  true  "Update profile payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/me [put]
func (h *PatientHandler) UpdateMe(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req UpdatePatientProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateProfile(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "profile updated", res)
}

// AssignPuskesmas handles PATCH /api/v1/admin/patients/:id/assign
// @Summary      Assign patient to puskesmas
// @Tags         patient
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                  true  "Patient ID"
// @Param        body  body  AssignPuskesmasRequest  true  "Puskesmas ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/patients/{id}/assign [patch]
func (h *PatientHandler) AssignPuskesmas(c fiber.Ctx) error {
	id := c.Params("id")
	var req AssignPuskesmasRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.AssignPuskesmas(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "puskesmas assigned successfully", res)
}

// ToggleStatus handles PATCH /api/v1/admin/patients/:id/status
// @Summary      Toggle patient account status
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Patient ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/patients/{id}/status [patch]
func (h *PatientHandler) ToggleStatus(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.ToggleStatus(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "patient status toggled", res)
}

// Delete handles DELETE /api/v1/admin/patients/:id
// @Summary      Delete patient account (soft-delete)
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Patient ID"
// @Success      204
// @Router       /admin/patients/{id} [delete]
func (h *PatientHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeletePatient(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}
