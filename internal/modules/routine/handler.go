package routine

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type RoutineHandler struct {
	svc RoutineService
	log *zap.Logger
}

func NewRoutineHandler(svc RoutineService, log *zap.Logger) *RoutineHandler {
	return &RoutineHandler{svc: svc, log: log}
}

// List handles GET /api/v1/patient/routines
// @Summary      Get routine setup
// @Tags         routine
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/routines [get]
func (h *RoutineHandler) List(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.ListRoutines(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "routines retrieved", items)
}

// Configure handles PUT /api/v1/patient/routines/:routineTimeId
// @Summary      Configure routine slot time and reminder
// @Tags         routine
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        routineTimeId  path  string                    true  "Routine Time ID"
// @Param        body           body  UpdateRoutineTimeRequest  true  "Configure payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/routines/{routineTimeId} [put]
func (h *RoutineHandler) Configure(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	routineTimeID := c.Params("routineTimeId")
	var req UpdateRoutineTimeRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.ConfigureRoutineTime(c.Context(), claims.UserID, routineTimeID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "routine time configured", res)
}

// Log handles POST /api/v1/patient/routines/log
// @Summary      Log routine execution
// @Tags         routine
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  LogRoutineRequest  true  "Logging payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/routines/log [post]
func (h *RoutineHandler) Log(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req LogRoutineRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.LogRoutine(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Success(c, "routine tracked successfully", res)
}

// Status handles GET /api/v1/patient/routines/status
// @Summary      Get routine onboarding status
// @Tags         routine
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/routines/status [get]
func (h *RoutineHandler) Status(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetOnboardingStatus(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "onboarding status checked", res)
}

// GetPatientActivityLogs handles GET /api/v1/admin/patients/:id/activities or /api/v1/staff/patients/:id/activities
func (h *RoutineHandler) GetPatientActivityLogs(c fiber.Ctx) error {
	patientID := c.Params("id")
	dateStr := c.Query("date")
	res, err := h.svc.GetPatientActivityLogs(c.Context(), patientID, dateStr)
	if err != nil {
		return err
	}
	return response.Success(c, "patient routine activity logs retrieved", res)
}
