package ai_chat

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type AIChatHandler struct {
	svc AIChatService
	log *zap.Logger
}

func NewAIChatHandler(svc AIChatService, log *zap.Logger) *AIChatHandler {
	return &AIChatHandler{
		svc: svc,
		log: log,
	}
}

// SendMessage handles POST /api/v1/ai/chat
func (h *AIChatHandler) SendMessage(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	patientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	var req SendMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request payload")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SendMessage(c.Context(), patientID, req)
	if err != nil {
		return errs.NewInternal("failed to process chat request: " + err.Error())
	}

	return response.Success(c, "assistant response generated", res)
}

// ListConversations handles GET /api/v1/ai/conversations
func (h *AIChatHandler) ListConversations(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	patientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetConversations(c.Context(), patientID)
	if err != nil {
		return errs.NewInternal("failed to fetch conversations")
	}

	return response.Success(c, "conversations retrieved successfully", res)
}

// CreateConversation handles POST /api/v1/ai/conversations
func (h *AIChatHandler) CreateConversation(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	patientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	var req CreateConversationRequest
	if err := c.Bind().Body(&req); err != nil {
		// Default title if empty body
		req.Title = "Percakapan Baru"
	}

	res, err := h.svc.CreateConversation(c.Context(), patientID, req.Title)
	if err != nil {
		return errs.NewInternal("failed to create conversation")
	}

	return response.Created(c, "conversation created successfully", res)
}

// GetMessages handles GET /api/v1/ai/conversations/:id/messages
func (h *AIChatHandler) GetMessages(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	patientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return errs.NewBadRequest("invalid conversation id format")
	}

	res, err := h.svc.GetMessages(c.Context(), patientID, convID)
	if err != nil {
		return errs.NewNotFound(err.Error())
	}

	return response.Success(c, "messages retrieved successfully", res)
}

// DeleteConversation handles DELETE /api/v1/ai/conversations/:id
func (h *AIChatHandler) DeleteConversation(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	patientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return errs.NewBadRequest("invalid conversation id format")
	}

	if err := h.svc.DeleteConversation(c.Context(), patientID, convID); err != nil {
		return errs.NewInternal("failed to delete conversation")
	}

	return response.Success(c, "conversation deleted successfully", nil)
}
