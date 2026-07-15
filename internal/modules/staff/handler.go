package staff

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type StaffHandler struct {
	svc StaffService
	log *zap.Logger
}

func NewStaffHandler(svc StaffService, log *zap.Logger) *StaffHandler {
	return &StaffHandler{svc: svc, log: log}
}

// List handles GET /api/v1/admin/staff
// @Summary      List all staff
// @Description  Get a paginated list of staff members (admin & staff)
// @Tags         staff
// @Security     BearerAuth
// @Produce      json
// @Param        role   query  string  false  "Filter by role (admin/staff)"
// @Param        page   query  int     false  "Page number (default: 1)"
// @Param        limit  query  int     false  "Limit (default: 10)"
// @Success      200    {object}  map[string]any
// @Router       /admin/staff [get]
func (h *StaffHandler) List(c fiber.Ctx) error {
	var r *domain.StaffRole
	roleStr := c.Query("role")
	if roleStr != "" {
		roleVal := domain.StaffRole(roleStr)
		r = &roleVal
	}

	search := c.Query("search")
	status := c.Query("status")

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

	items, total, err := h.svc.ListStaff(c.Context(), search, status, r, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "staff retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/admin/staff/:id
// @Summary      Get staff by ID
// @Tags         staff
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Staff ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/staff/{id} [get]
func (h *StaffHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.GetStaff(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "staff retrieved", res)
}

// Create handles POST /api/v1/admin/staff
// @Summary      Create staff account
// @Tags         staff
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateStaffRequest  true  "Create staff payload"
// @Success      201  {object}  map[string]any
// @Router       /admin/staff [post]
func (h *StaffHandler) Create(c fiber.Ctx) error {
	var req CreateStaffRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateStaff(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "staff account created", res)
}

// Update handles PUT /api/v1/admin/staff/:id
// @Summary      Update staff account
// @Tags         staff
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string              true  "Staff ID"
// @Param        body  body  UpdateStaffRequest  true  "Update staff payload"
// @Success      200  {object}  map[string]any
// @Router       /admin/staff/{id} [put]
func (h *StaffHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateStaffRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateStaff(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "staff account updated", res)
}

// ToggleStatus handles PATCH /api/v1/admin/staff/:id/status
// @Summary      Toggle staff active/inactive status
// @Tags         staff
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Staff ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/staff/{id}/status [patch]
func (h *StaffHandler) ToggleStatus(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.ToggleStatus(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "staff status toggled", res)
}

// GetMe handles GET /api/v1/admin/me
// @Summary      Get current staff profile
// @Tags         staff
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /admin/me [get]
func (h *StaffHandler) GetMe(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetStaff(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "profile retrieved", res)
}

// UpdateMe handles PUT /api/v1/admin/me
// @Summary      Update current staff profile
// @Tags         staff
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  UpdateProfileRequest  true  "Update profile payload"
// @Success      200  {object}  map[string]any
// @Router       /admin/me [put]
func (h *StaffHandler) UpdateMe(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateMyProfile(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "profile updated", res)
}

// ChangePassword handles PUT /api/v1/admin/me/password
// @Summary      Change current staff password
// @Tags         staff
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  ChangePasswordRequest  true  "Change password payload"
// @Success      200  {object}  map[string]any
// @Router       /admin/me/password [put]
func (h *StaffHandler) ChangePassword(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	err := h.svc.ChangePassword(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "password updated successfully", nil)
}

// Delete handles DELETE /api/v1/admin/staff/:id
// @Summary      Delete staff account (soft-delete)
// @Tags         staff
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Staff ID"
// @Success      204
// @Router       /admin/staff/{id} [delete]
func (h *StaffHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteStaff(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}
