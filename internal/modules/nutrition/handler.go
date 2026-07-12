package nutrition

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type NutritionHandler struct {
	svc NutritionService
	log *zap.Logger
}

func NewNutritionHandler(svc NutritionService, log *zap.Logger) *NutritionHandler {
	return &NutritionHandler{svc: svc, log: log}
}

// Search handles GET /api/v1/foods
// @Summary      Search foods
// @Tags         nutrition
// @Security     BearerAuth
// @Produce      json
// @Param        q     query  string  true  "Search query"
// @Success      200  {object}  map[string]any
// @Router       /foods [get]
func (h *NutritionHandler) Search(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	query := c.Query("q")
	items, err := h.svc.SearchFoods(c.Context(), claims.UserID, query)
	if err != nil {
		return err
	}
	return response.Success(c, "foods retrieved", items)
}

// GetRecent handles GET /api/v1/foods/recent
// @Summary      Get recent food searches
// @Tags         nutrition
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /foods/recent [get]
func (h *NutritionHandler) GetRecent(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.GetRecentFoods(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "recent foods retrieved", items)
}

// LogMeal handles POST /api/v1/patient/meals
// @Summary      Log meal consumption
// @Tags         nutrition
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  LogMealRequest  true  "Log meal payload"
// @Success      201  {object}  map[string]any
// @Router       /patient/meals [post]
func (h *NutritionHandler) LogMeal(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req LogMealRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.LogMeal(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "meal logged successfully", res)
}

// GetSummary handles GET /api/v1/patient/meals/summary
// @Summary      Get daily calorie and nutrition summary
// @Tags         nutrition
// @Security     BearerAuth
// @Produce      json
// @Param        date  query  string  false  "Date (YYYY-MM-DD, default: today)"
// @Success      200  {object}  map[string]any
// @Router       /patient/meals/summary [get]
func (h *NutritionHandler) GetSummary(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	dateStr := c.Query("date")
	res, err := h.svc.GetDailyNutritionSummary(c.Context(), claims.UserID, dateStr)
	if err != nil {
		return err
	}
	return response.Success(c, "nutrition summary retrieved", res)
}

// CreateFood handles POST /api/v1/admin/foods
// @Summary      Create global food entry (Admin)
// @Tags         nutrition
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateFoodRequest  true  "Create food payload"
// @Success      201  {object}  map[string]any
// @Router       /admin/foods [post]
func (h *NutritionHandler) CreateFood(c fiber.Ctx) error {
	var req CreateFoodRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateFood(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Created(c, "food item created", res)
}

// UpdateFood handles PUT /api/v1/admin/foods/:id
// @Summary      Update global food entry (Admin)
// @Tags         nutrition
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string             true  "Food ID"
// @Param        body  body  CreateFoodRequest  true  "Update food payload"
// @Success      200  {object}  map[string]any
// @Router       /admin/foods/{id} [put]
func (h *NutritionHandler) UpdateFood(c fiber.Ctx) error {
	id := c.Params("id")
	var req CreateFoodRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateFood(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "food item updated", res)
}

// GetPatientMealLogs handles GET /api/v1/admin/patients/:id/meals or /api/v1/staff/patients/:id/meals
func (h *NutritionHandler) GetPatientMealLogs(c fiber.Ctx) error {
	patientID := c.Params("id")
	dateStr := c.Query("date")
	res, err := h.svc.GetPatientMealLogs(c.Context(), patientID, dateStr)
	if err != nil {
		return err
	}
	return response.Success(c, "patient meal logs retrieved", res)
}
