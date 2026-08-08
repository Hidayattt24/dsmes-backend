package reminder

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type ReminderHandler struct {
	svc ReminderService
	log *zap.Logger
}

func NewReminderHandler(svc ReminderService, log *zap.Logger) *ReminderHandler {
	return &ReminderHandler{svc: svc, log: log}
}

// List handles GET /api/v1/patient/reminders
// @Summary      Get list of reminders
// @Tags         reminder
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/reminders [get]
func (h *ReminderHandler) List(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.ListReminders(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "reminders retrieved", items)
}

// Create handles POST /api/v1/patient/reminders
// @Summary      Create personal reminder
// @Tags         reminder
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateReminderRequest  true  "Create reminder payload"
// @Success      201  {object}  map[string]any
// @Router       /patient/reminders [post]
func (h *ReminderHandler) Create(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req CreateReminderRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateReminder(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "reminder created", res)
}

// Update handles PUT /api/v1/patient/reminders/:id
// @Summary      Update personal reminder
// @Tags         reminder
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                 true  "Reminder ID"
// @Param        body  body  CreateReminderRequest  true  "Update reminder payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/reminders/{id} [put]
func (h *ReminderHandler) Update(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	var req CreateReminderRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateReminder(c.Context(), claims.UserID, id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "reminder updated", res)
}

// Toggle handles PATCH /api/v1/patient/reminders/:id/toggle
// @Summary      Toggle reminder status
// @Tags         reminder
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Reminder ID"
// @Success      200  {object}  map[string]any
// @Router       /patient/reminders/{id}/toggle [patch]
func (h *ReminderHandler) Toggle(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	res, err := h.svc.ToggleReminder(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}
	return response.Success(c, "reminder status toggled", res)
}

// Delete handles DELETE /api/v1/patient/reminders/:id
// @Summary      Delete reminder
// @Tags         reminder
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Reminder ID"
// @Success      204
// @Router       /patient/reminders/{id} [delete]
func (h *ReminderHandler) Delete(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	if err := h.svc.DeleteReminder(c.Context(), claims.UserID, id); err != nil {
		return err
	}
	return response.NoContent(c)
}

// GetNotifications handles GET /api/v1/patient/notifications
// @Summary      Get patient notifications
// @Tags         reminder
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/notifications [get]
func (h *ReminderHandler) GetNotifications(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.GetNotifications(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "notifications retrieved", items)
}

// MarkRead handles PATCH /api/v1/patient/notifications/read
// @Summary      Mark all notifications as read
// @Tags         reminder
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/notifications/read [patch]
func (h *ReminderHandler) MarkRead(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	if err := h.svc.MarkAllRead(c.Context(), claims.UserID); err != nil {
		return err
	}
	return response.Success(c, "all notifications marked as read", nil)
}

func (h *ReminderHandler) MarkNotificationReadByID(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	if err := h.svc.MarkReadByID(c.Context(), claims.UserID, id); err != nil {
		return err
	}
	return response.Success(c, "notification marked as read", nil)
}

func (h *ReminderHandler) DeleteNotificationByID(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	if err := h.svc.DeleteNotificationByID(c.Context(), claims.UserID, id); err != nil {
		return err
	}
	return response.NoContent(c)
}

// LogMedication handles POST /api/v1/patient/medications/log
// @Summary      Log medication as taken/skipped for today
// @Tags         reminder
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  LogMedicationRequest  true  "Medication log payload"
// @Success      200  {object}  map[string]any
// @Router       /patient/medications/log [post]
func (h *ReminderHandler) LogMedication(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req LogMedicationRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.LogMedication(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "medication logged successfully", res)
}

// GetPatientMedicationLogs handles GET /api/v1/admin/patients/:id/medications or /api/v1/staff/patients/:id/medications
func (h *ReminderHandler) GetPatientMedicationLogs(c fiber.Ctx) error {
	patientID := c.Params("id")
	dateStr := c.Query("date")
	res, err := h.svc.GetPatientMedicationLogs(c.Context(), patientID, dateStr)
	if err != nil {
		return err
	}
	return response.Success(c, "patient medication logs retrieved", res)
}
