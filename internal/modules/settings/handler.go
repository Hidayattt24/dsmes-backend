package settings

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type SettingsHandler struct {
	svc SettingsService
	log *zap.Logger
}

func NewSettingsHandler(svc SettingsService, log *zap.Logger) *SettingsHandler {
	return &SettingsHandler{svc: svc, log: log}
}

// GetFAQs handles GET /api/v1/faqs
// @Summary      Get list of active FAQs
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /faqs [get]
func (h *SettingsHandler) GetFAQs(c fiber.Ctx) error {
	items, err := h.svc.GetFAQs(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "FAQs retrieved", items)
}

// SubmitTicket handles POST /api/v1/patient/support/tickets
// @Summary      Submit a support ticket
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateTicketRequest  true  "Ticket payload"
// @Success      201  {object}  map[string]any
// @Router       /patient/support/tickets [post]
func (h *SettingsHandler) SubmitTicket(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req CreateTicketRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SubmitTicket(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "support ticket submitted", res)
}

// GetMyTickets handles GET /api/v1/patient/support/tickets
// @Summary      Get patient support tickets
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/support/tickets [get]
func (h *SettingsHandler) GetMyTickets(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.GetPatientTickets(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "support tickets retrieved", items)
}

// GetAllTickets handles GET /api/v1/admin/support/tickets
// @Summary      Get all support tickets (Admin)
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /admin/support/tickets [get]
func (h *SettingsHandler) GetAllTickets(c fiber.Ctx) error {
	items, err := h.svc.GetAllTickets(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "all support tickets retrieved", items)
}

// ResolveTicket handles PATCH /api/v1/admin/support/tickets/:id/resolve
// @Summary      Mark support ticket as resolved (Admin)
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Ticket ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/support/tickets/{id}/resolve [patch]
func (h *SettingsHandler) ResolveTicket(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.ResolveTicket(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "support ticket resolved", res)
}
