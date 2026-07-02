package checkin

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type CheckinHandler struct {
	svc CheckinService
	log *zap.Logger
}

func NewCheckinHandler(svc CheckinService, log *zap.Logger) *CheckinHandler {
	return &CheckinHandler{svc: svc, log: log}
}

// Checkin handles POST /api/v1/patient/checkin
// @Summary      Perform daily check-in
// @Tags         checkin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CheckinRequest  true  "Check-in payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/checkin [post]
func (h *CheckinHandler) Checkin(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req CheckinRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.PerformCheckin(c.Context(), claims.UserID, req.CheckinDate)
	if err != nil {
		return err
	}
	return response.Success(c, "daily check-in completed", res)
}

// GetCalendar handles GET /api/v1/patient/checkin/calendar
// @Summary      Get monthly check-in calendar
// @Tags         checkin
// @Security     BearerAuth
// @Produce      json
// @Param        year   query  int  false  "Year (default: current year)"
// @Param        month  query  int  false  "Month (default: current month)"
// @Success      200  {object}  map[string]any
// @Router       /patient/checkin/calendar [get]
func (h *CheckinHandler) GetCalendar(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	now := time.Now()
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if val, err := strconv.Atoi(yearStr); err == nil {
			year = val
		}
	}
	if monthStr != "" {
		if val, err := strconv.Atoi(monthStr); err == nil {
			month = val
		}
	}

	res, err := h.svc.GetCheckinCalendar(c.Context(), claims.UserID, year, month)
	if err != nil {
		return err
	}
	return response.Success(c, "check-in calendar retrieved", res)
}
