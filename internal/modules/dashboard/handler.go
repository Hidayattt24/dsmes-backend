package dashboard

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
)

type DashboardHandler struct {
	svc DashboardService
	log *zap.Logger
}

func NewDashboardHandler(svc DashboardService, log *zap.Logger) *DashboardHandler {
	return &DashboardHandler{svc: svc, log: log}
}

// GetAdmin handles GET /api/v1/admin/dashboard/stats
func (h *DashboardHandler) GetAdmin(c fiber.Ctx) error {
	res, err := h.svc.GetAdminDashboard(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "admin dashboard stats retrieved", res)
}

// GetStaff handles GET /api/v1/staff/dashboard/stats
func (h *DashboardHandler) GetStaff(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	res, err := h.svc.GetStaffDashboard(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "staff dashboard stats retrieved", res)
}
