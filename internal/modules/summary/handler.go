package summary

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type SummaryHandler struct {
	svc SummaryService
	log *zap.Logger
}

func NewSummaryHandler(svc SummaryService, log *zap.Logger) *SummaryHandler {
	return &SummaryHandler{svc: svc, log: log}
}

// GetLatest handles GET /api/v1/patient/summary/weekly
// @Summary      Get latest weekly summary
// @Tags         summary
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/summary/weekly [get]
func (h *SummaryHandler) GetLatest(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetLatestSummary(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "latest weekly summary retrieved", res)
}

// Generate handles POST /api/v1/internal/summary/generate
// @Summary      Generate weekly summary (Internal/Cron)
// @Tags         summary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  GenerateSummaryRequest  true  "Generate payload"
// @Success      200  {object}  map[string]any
// @Router       /internal/summary/generate [post]
func (h *SummaryHandler) Generate(c fiber.Ctx) error {
	var req GenerateSummaryRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.GenerateWeeklySummary(c.Context(), req.PatientID, req.WeekStartDate)
	if err != nil {
		return err
	}
	return response.Success(c, "weekly summary generated successfully", res)
}
