package survey

import (
	"fmt"
	"strconv"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type SurveyHandler struct {
	svc SurveyService
	log *zap.Logger
}

func NewSurveyHandler(svc SurveyService, log *zap.Logger) *SurveyHandler {
	return &SurveyHandler{svc: svc, log: log}
}

func (h *SurveyHandler) Create(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return errs.NewUnauthorized("unauthorized access")
	}

	var req CreateSurveyRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateSurvey(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "survey created successfully", res)
}

func (h *SurveyHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	var req UpdateSurveyRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateSurvey(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "survey updated successfully", res)
}

func (h *SurveyHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	if err := h.svc.DeleteSurvey(c.Context(), id); err != nil {
		return err
	}
	return response.Success(c, "survey deleted successfully", nil)
}

func (h *SurveyHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	claims := middleware.ClaimsFromContext(c)
	isPatient := claims != nil && claims.Role == "user"

	res, err := h.svc.GetSurveyByID(c.Context(), id, isPatient)
	if err != nil {
		return err
	}
	return response.Success(c, "survey detail fetched successfully", res)
}

func (h *SurveyHandler) List(c fiber.Ctx) error {
	surveyType := c.Query("type")
	status := c.Query("status")
	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 20
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	items, total, err := h.svc.ListSurveys(c.Context(), surveyType, status, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "surveys fetched successfully", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *SurveyHandler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	var req UpdateSurveyStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateStatus(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "survey status updated successfully", res)
}

func (h *SurveyHandler) Duplicate(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return errs.NewUnauthorized("unauthorized access")
	}

	res, err := h.svc.DuplicateSurvey(c.Context(), id, claims.UserID)
	if err != nil {
		return err
	}
	return response.Created(c, "survey duplicated successfully", res)
}

func (h *SurveyHandler) GetActivePatientSurvey(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return errs.NewUnauthorized("unauthorized access")
	}

	surveyType := c.Query("type")

	res, err := h.svc.GetActiveSurveysForPatient(c.Context(), surveyType, claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "active surveys fetched successfully", res)
}

func (h *SurveyHandler) Submit(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return errs.NewUnauthorized("unauthorized access")
	}

	var req SubmitSurveyRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SubmitSurvey(c.Context(), id, claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "survey submitted successfully", res)
}

func (h *SurveyHandler) GetResponses(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 20
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	items, total, err := h.svc.GetSurveyResponses(c.Context(), id, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "survey responses fetched successfully", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *SurveyHandler) GetAnalytics(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	res, err := h.svc.GetSurveyAnalytics(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "survey analytics fetched successfully", res)
}

func (h *SurveyHandler) ExportCSV(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("survey ID is required")
	}

	data, filename, err := h.svc.ExportResponsesCSV(c.Context(), id)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Send(data)
}
