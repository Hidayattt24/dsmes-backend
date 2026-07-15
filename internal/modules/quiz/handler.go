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

// List handles GET /api/v1/admin/quiz or /api/v1/staff/quiz
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
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	items, total, err := h.svc.ListQuizzes(c.Context(), search, status, sortBy, sortOrder, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "quizzes retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/admin/quiz/:id or /api/v1/staff/quiz/:id
func (h *QuizHandler) GetByID(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	isAdminOrStaff := claims != nil && (claims.Role == "admin" || claims.Role == "staff")

	id := c.Params("id")
	res, err := h.svc.GetQuiz(c.Context(), id, isAdminOrStaff)
	if err != nil {
		return err
	}
	return response.Success(c, "quiz details retrieved", res)
}

// Create handles POST /api/v1/admin/quiz
func (h *QuizHandler) Create(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req CreateQuizRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateQuiz(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "quiz created successfully", res)
}

// Update handles PUT /api/v1/admin/quiz/:id
func (h *QuizHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req CreateQuizRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateQuiz(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "quiz updated successfully", res)
}

// Delete handles DELETE /api/v1/admin/quiz/:id
func (h *QuizHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteQuiz(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}

// GetStats handles GET /api/v1/admin/quiz/stats or /api/v1/staff/quiz/stats
func (h *QuizHandler) GetStats(c fiber.Ctx) error {
	res, err := h.svc.GetStats(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "quiz statistics retrieved", res)
}

// ListParticipants handles GET /api/v1/admin/quiz/:id/participants or /api/v1/staff/quiz/:id/participants
func (h *QuizHandler) ListParticipants(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.ListParticipants(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "quiz participants retrieved", res)
}

// GetParticipantDetail handles GET /api/v1/admin/quiz/:id/participant/:participant_id or /api/v1/staff/quiz/:id/participant/:participant_id
func (h *QuizHandler) GetParticipantDetail(c fiber.Ctx) error {
	id := c.Params("id")
	participantID := c.Params("participant_id")
	res, err := h.svc.GetParticipantDetail(c.Context(), id, participantID)
	if err != nil {
		return err
	}
	return response.Success(c, "participant details retrieved", res)
}
