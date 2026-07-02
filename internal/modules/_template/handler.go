package _template

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

// TemplateHandler handles HTTP requests for the Template module.
//
// Rules:
//   - Never import gorm or sql here
//   - Never write business logic here
//   - Responsibilities: parse request → validate → call service → write response
type TemplateHandler struct {
	svc TemplateService
	log *zap.Logger
}

// NewTemplateHandler creates a handler instance with the given service.
// Called from routes.go when wiring the module.
func NewTemplateHandler(svc TemplateService, log *zap.Logger) *TemplateHandler {
	return &TemplateHandler{svc: svc, log: log}
}

// List handles GET /api/v1/templates
// @Summary      List templates
// @Description  Returns a paginated list of templates
// @Tags         templates
// @Produce      json
// @Param        page   query  int  false  "Page number (default: 1)"
// @Param        limit  query  int  false  "Items per page (default: 20, max: 100)"
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /templates [get]
func (h *TemplateHandler) List(c fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	items, total, err := h.svc.List(c.Context(), page, limit)
	if err != nil {
		return err // Fiber's ErrorHandler resolves *errs.AppError automatically
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "templates retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/templates/:id
// @Summary      Get template by ID
// @Tags         templates
// @Produce      json
// @Param        id  path  string  true  "Template UUID"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /templates/{id} [get]
func (h *TemplateHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "template retrieved", item)
}

// Create handles POST /api/v1/templates
// @Summary      Create template
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        body  body  TemplateRequest  true  "Template data"
// @Success      201  {object}  map[string]any
// @Failure      422  {object}  map[string]any
// @Router       /templates [post]
func (h *TemplateHandler) Create(c fiber.Ctx) error {
	var req TemplateRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	item, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "template created", item)
}

// Update handles PUT /api/v1/templates/:id
// @Summary      Update template
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id    path  string           true  "Template UUID"
// @Param        body  body  TemplateRequest  true  "Template data"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Failure      422  {object}  map[string]any
// @Router       /templates/{id} [put]
func (h *TemplateHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")

	var req TemplateRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	item, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "template updated", item)
}

// Delete handles DELETE /api/v1/templates/:id
// @Summary      Delete template
// @Tags         templates
// @Produce      json
// @Param        id  path  string  true  "Template UUID"
// @Success      204
// @Failure      404  {object}  map[string]any
// @Router       /templates/{id} [delete]
func (h *TemplateHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}
