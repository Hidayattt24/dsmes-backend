package history

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
)

type HistoryHandler struct {
	svc HistoryService
	log *zap.Logger
}

func NewHistoryHandler(svc HistoryService, log *zap.Logger) *HistoryHandler {
	return &HistoryHandler{svc: svc, log: log}
}

// GetPatientHistory handles GET /api/v1/patient/history
// @Summary      Get patient activity history
// @Description  Returns a paginated timeline of all patient activities (blood sugar, meals, activities, medications, measurements)
// @Tags         history
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default: 1)"
// @Param        limit  query  int  false  "Items per page (default: 50, max: 100)"
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /patient/history [get]
func (h *HistoryHandler) GetPatientHistory(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	items, total, err := h.svc.GetPatientHistory(c.Context(), claims.UserID, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "patient history retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// DeleteHistoryItem handles DELETE /api/v1/patient/history/:type/:id
func (h *HistoryHandler) DeleteHistoryItem(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}
	activityType := c.Params("type")
	id := c.Params("id")

	if err := h.svc.DeleteHistoryItem(c.Context(), claims.UserID, activityType, id); err != nil {
		return err
	}
	return response.NoContent(c)
}
