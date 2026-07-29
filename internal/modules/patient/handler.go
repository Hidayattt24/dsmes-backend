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

// SetupHealthProfile handles POST /api/v1/patient/profile/setup
// @Summary      Setup patient health profile
// @Tags         patient
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  SetupHealthProfileRequest  true  "Health profile setup payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/profile/setup [post]
func (h *PatientHandler) SetupHealthProfile(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req SetupHealthProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SetupHealthProfile(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "health profile set up successfully", res)
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
	gender := c.Query("gender")
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")
	bloodSugarStatus := c.Query("blood_sugar_status")
	riskLevel := c.Query("risk_level")

	complianceMin := (*int)(nil)
	complianceMax := (*int)(nil)
	if cmStr := c.Query("compliance_min"); cmStr != "" {
		if cm, err := strconv.Atoi(cmStr); err == nil {
			complianceMin = &cm
		}
	}
	if cmxStr := c.Query("compliance_max"); cmxStr != "" {
		if cmx, err := strconv.Atoi(cmxStr); err == nil {
			complianceMax = &cmx
		}
	}

	items, total, err := h.svc.ListPatients(c.Context(), PatientFilterQuery{
		Search:           search,
		Gender:           gender,
		Status:           status,
		SortBy:           sortBy,
		SortOrder:        sortOrder,
		BloodSugarStatus: bloodSugarStatus,
		RiskLevel:        riskLevel,
		ComplianceMin:    complianceMin,
		ComplianceMax:    complianceMax,
		Page:             page,
		Limit:            limit,
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

// ListStaff handles GET /api/v1/staff/patients
// @Summary      List assigned patients (Staff)
// @Description  Returns paginated list of assigned patients with optional filters. Staff ID is auto-inferred from JWT claims.
// @Tags         patient
// @Security     BearerAuth
// @Produce      json
// @Param        search              query  string  false  "Search pattern (name/email)"
// @Param        gender              query  string  false  "Gender filter (Laki-laki/Perempuan)"
// @Param        status              query  string  false  "Account status filter (Aktif/Nonaktif)"
// @Param        page                query  int     false  "Page number (default: 1)"
// @Param        limit               query  int     false  "Items per page (default: 10)"
// @Param        sort_by             query  string  false  "Sort field (name, newest, oldest, latest_record, highest_blood_sugar)"
// @Param        sort_order          query  string  false  "Sort direction (asc/desc)"
// @Param        blood_sugar_status  query  string  false  "Blood sugar status filter (normal, tinggi, rendah, sangat_tinggi)"
// @Param        risk_level          query  string  false  "Risk level filter (rendah, sedang, tinggi, sangat_tinggi)"
// @Param        compliance_min      query  int     false  "Minimum compliance percentage"
// @Param        compliance_max      query  int     false  "Maximum compliance percentage"
// @Param        age_min             query  int     false  "Minimum age in years"
// @Param        age_max             query  int     false  "Maximum age in years"
// @Success      200                 {object}  map[string]any
// @Router       /staff/patients [get]
func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}

func (h *PatientHandler) ListStaff(c fiber.Ctx) error {
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

	items, total, err := h.svc.ListPatients(c.Context(), PatientFilterQuery{
		StaffID:          claims.UserID,
		Search:           c.Query("search"),
		Gender:           c.Query("gender"),
		Status:           c.Query("status"),
		SortBy:           c.Query("sort_by"),
		SortOrder:        c.Query("sort_order"),
		BloodSugarStatus: c.Query("blood_sugar_status"),
		RiskLevel:        c.Query("risk_level"),
		ComplianceMin:    parseIntPtr(c.Query("compliance_min")),
		ComplianceMax:    parseIntPtr(c.Query("compliance_max")),
		AgeMin:           parseIntPtr(c.Query("age_min")),
		AgeMax:           parseIntPtr(c.Query("age_max")),
		Page:             page,
		Limit:            limit,
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

// GetByID handles GET /api/v1/admin/patients/:id or /api/v1/staff/patients/:id
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

// UpdatePatientByAdmin handles PUT /api/v1/admin/patients/:id
func (h *PatientHandler) UpdatePatientByAdmin(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdatePatientRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid payload", err)
	}

	res, err := h.svc.UpdatePatientByAdmin(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "patient profile updated by admin", res)
}

// CreateMeasurement handles POST /api/v1/admin/patients/:id/measurements or POST /api/v1/patient/measurements
func (h *PatientHandler) CreateMeasurement(c fiber.Ctx) error {
	patientID := c.Params("id")
	claims := middleware.ClaimsFromContext(c)

	recByID := ""
	recByName := "Admin"
	recByRole := "admin"

	if claims != nil {
		recByID = claims.UserID
		recByRole = claims.Role
		if claims.Role == "patient" && patientID == "" {
			patientID = claims.UserID
			recByName = "Pasien"
		} else if claims.Role == "admin" {
			recByName = "Admin"
		}
	}

	var req CreateMeasurementRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid payload", err)
	}

	res, err := h.svc.CreateMeasurement(c.Context(), patientID, req, recByID, recByName, recByRole)
	if err != nil {
		return err
	}
	return response.Created(c, "health measurement record created", res)
}

// GetPatientMeasurements handles GET /api/v1/admin/patients/:id/measurements or /api/v1/staff/patients/:id/measurements
func (h *PatientHandler) GetPatientMeasurements(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		claims := middleware.ClaimsFromContext(c)
		if claims != nil {
			id = claims.UserID
		}
	}

	res, err := h.svc.GetPatientMeasurements(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "patient measurements history retrieved", res)
}

// UpdateMeasurement handles PUT /api/v1/admin/patients/:id/measurements/:measurementId
func (h *PatientHandler) UpdateMeasurement(c fiber.Ctx) error {
	patientID := c.Params("id")
	measurementID := c.Params("measurementId")

	var req UpdateMeasurementRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid payload", err)
	}

	res, err := h.svc.UpdateMeasurement(c.Context(), patientID, measurementID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "measurement record updated", res)
}

// GetPatientActivityAnalytics handles GET /api/v1/admin/patients/:id/activity-analytics
func (h *PatientHandler) GetPatientActivityAnalytics(c fiber.Ctx) error {
	id := c.Params("id")
	days, _ := strconv.Atoi(c.Query("days", "7"))
	if days < 0 {
		days = 7
	}
	res, err := h.svc.GetPatientActivityAnalytics(c.Context(), id, days)
	if err != nil {
		return err
	}
	return response.Success(c, "patient activity analytics retrieved", res)
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

// ChangePassword handles PUT /api/v1/patient/me/password
// @Summary      Change current patient password
// @Tags         patient
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  ChangePasswordRequest  true  "Change password payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/me/password [put]
func (h *PatientHandler) ChangePassword(c fiber.Ctx) error {
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

	if err := h.svc.ChangePassword(c.Context(), claims.UserID, req); err != nil {
		return err
	}

	return response.Success(c, "password changed", nil)
}

// AssignStaff handles PATCH /api/v1/admin/patients/:id/assign
// @Summary      Assign patient to monitoring staff
// @Tags         patient
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string              true  "Patient ID"
// @Param        body  body  AssignStaffRequest  true  "Staff ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/patients/{id}/assign [patch]
func (h *PatientHandler) AssignStaff(c fiber.Ctx) error {
	id := c.Params("id")
	var req AssignStaffRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.AssignStaff(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "staff assigned successfully", res)
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

// GetStats handles GET /api/v1/admin/patients/stats
func (h *PatientHandler) GetStats(c fiber.Ctx) error {
	res, err := h.svc.GetStats(c.Context(), "")
	if err != nil {
		return err
	}
	return response.Success(c, "patient statistics retrieved", res)
}

// GetStatsStaff handles GET /api/v1/staff/patients/stats
func (h *PatientHandler) GetStatsStaff(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	res, err := h.svc.GetStats(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "patient statistics retrieved", res)
}
