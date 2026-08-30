package quiz

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type QuizHandler struct {
	svc QuizService
	log *zap.Logger
}

func NewQuizHandler(svc QuizService, log *zap.Logger) *QuizHandler {
	return &QuizHandler{svc: svc, log: log}
}

func (h *QuizHandler) List(c fiber.Ctx) error {
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
	qType := c.Query("type")
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	claims := middleware.ClaimsFromContext(c)
	var items []QuestionnaireResponse
	var total int64
	var err error
	if claims != nil && claims.Role == "staff" {
		items, total, err = h.svc.ListQuestionnairesForStaff(c.Context(), claims.UserID, search, qType, status, sortBy, sortOrder, page, limit)
	} else {
		items, total, err = h.svc.ListQuestionnaires(c.Context(), search, qType, status, sortBy, sortOrder, page, limit)
	}
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "questionnaires retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *QuizHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	claims := middleware.ClaimsFromContext(c)
	var item *QuestionnaireDetailResponse
	var err error
	if claims != nil && claims.Role == "staff" {
		item, err = h.svc.GetQuestionnaireForStaff(c.Context(), claims.UserID, id)
	} else {
		isAdminOrStaff := claims != nil && (claims.Role == "admin" || claims.Role == "staff")
		item, err = h.svc.GetQuestionnaire(c.Context(), id, isAdminOrStaff)
	}
	if err != nil {
		return err
	}

	return response.Success(c, "questionnaire detail retrieved", item)
}

func (h *QuizHandler) GetActivePreTest(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	isAdminOrStaff := claims != nil && (claims.Role == "admin" || claims.Role == "staff")

	item, err := h.svc.GetActivePreTest(c.Context(), isAdminOrStaff)
	if err != nil {
		return err
	}

	return response.Success(c, "active Pre-Test retrieved", item)
}

func (h *QuizHandler) GetPostTestByEducation(c fiber.Ctx) error {
	educationID := c.Query("education_id")
	if educationID == "" {
		return errs.NewBadRequest("education_id parameter is required")
	}

	claims := middleware.ClaimsFromContext(c)
	isAdminOrStaff := claims != nil && (claims.Role == "admin" || claims.Role == "staff")

	item, err := h.svc.GetPostTestByEducation(c.Context(), educationID, isAdminOrStaff)
	if err != nil {
		return err
	}

	return response.Success(c, "Post-Test retrieved", item)
}

func (h *QuizHandler) Create(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	staffID := ""
	if claims != nil {
		staffID = claims.UserID
	}

	var req CreateQuestionnaireRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateQuestionnaire(c.Context(), staffID, req)
	if err != nil {
		return err
	}

	return response.Created(c, "questionnaire created successfully", res)
}

func (h *QuizHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	var req CreateQuestionnaireRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateQuestionnaire(c.Context(), id, req)
	if err != nil {
		return err
	}

	return response.Success(c, "questionnaire updated successfully", res)
}

func (h *QuizHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	if err := h.svc.DeleteQuestionnaire(c.Context(), id); err != nil {
		return err
	}

	return response.NoContent(c)
}

func (h *QuizHandler) GetStats(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	var stats *QuizStats
	var err error
	if claims != nil && claims.Role == "staff" {
		stats, err = h.svc.GetStatsForStaff(c.Context(), claims.UserID)
	} else {
		stats, err = h.svc.GetStats(c.Context())
	}
	if err != nil {
		return err
	}
	return response.Success(c, "questionnaire stats retrieved", stats)
}

func (h *QuizHandler) Submit(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	var req SubmitQuestionnaireRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SubmitQuestionnaire(c.Context(), claims.UserID, id, req)
	if err != nil {
		return err
	}

	return response.Created(c, "questionnaire submitted successfully", res)
}

func (h *QuizHandler) ListParticipants(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	claims := middleware.ClaimsFromContext(c)
	var participants []ParticipantResponse
	var err error
	if claims != nil && claims.Role == "staff" {
		participants, err = h.svc.ListParticipantsForStaff(c.Context(), claims.UserID, id)
	} else {
		participants, err = h.svc.ListParticipants(c.Context(), id)
	}
	if err != nil {
		return err
	}

	return response.Success(c, "questionnaire participants retrieved", participants)
}

func (h *QuizHandler) GetParticipantDetail(c fiber.Ctx) error {
	id := c.Params("id")
	participantID := c.Params("participant_id")
	if id == "" || participantID == "" {
		return errs.NewBadRequest("questionnaire ID and participant ID are required")
	}

	claims := middleware.ClaimsFromContext(c)
	var detail *ParticipantDetailResponse
	var err error
	if claims != nil && claims.Role == "staff" {
		detail, err = h.svc.GetParticipantDetailForStaff(c.Context(), claims.UserID, id, participantID)
	} else {
		detail, err = h.svc.GetParticipantDetail(c.Context(), id, participantID)
	}
	if err != nil {
		return err
	}

	return response.Success(c, "participant detail retrieved", detail)
}

func (h *QuizHandler) ListPatient(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	qType := c.Query("type")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	res, total, err := h.svc.ListPatientQuestionnaires(c.Context(), qType, claims.UserID, page, perPage)
	if err != nil {
		return err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "questionnaires retrieved", res, &response.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *QuizHandler) GetMyAttempt(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	res, err := h.svc.GetMyAttempt(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}

	return response.Success(c, "my attempt retrieved", res)
}

func (h *QuizHandler) GetMyAttemptDetail(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	id := c.Params("id")
	if id == "" {
		return errs.NewBadRequest("questionnaire ID is required")
	}

	res, err := h.svc.GetMyAttemptDetail(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}

	return response.Success(c, "my attempt detail retrieved", res)
}

func (h *QuizHandler) GetMyAttemptDetailByID(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}
	questionnaireID := c.Params("id")
	attemptID := c.Params("attempt_id")
	if questionnaireID == "" || attemptID == "" {
		return errs.NewBadRequest("questionnaire ID and attempt ID are required")
	}
	res, err := h.svc.GetMyAttemptDetailByID(c.Context(), claims.UserID, questionnaireID, attemptID)
	if err != nil {
		return err
	}
	return response.Success(c, "my attempt detail retrieved", res)
}

func (h *QuizHandler) GetMyHistory(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	qType := c.Query("type")

	res, err := h.svc.GetMyHistory(c.Context(), claims.UserID, qType)
	if err != nil {
		return err
	}

	return response.Success(c, "questionnaire history retrieved", res)
}
