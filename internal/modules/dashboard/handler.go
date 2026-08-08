package dashboard

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

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
	res, err := h.svc.GetStaffDashboard(c.Context(), "")
	if err != nil {
		return err
	}
	return response.Success(c, "staff dashboard stats retrieved", res)
}

// GetTopArticles handles GET /api/v1/admin/dashboard/top-articles
// @Summary      Get top read articles (Admin)
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /admin/dashboard/top-articles [get]
func (h *DashboardHandler) GetTopArticles(c fiber.Ctx) error {
	res, err := h.svc.GetTopArticles(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "top articles retrieved", res)
}

// GetActivityChart handles GET /api/v1/admin/dashboard/activity-chart
// @Summary      Get 7-day activity stats chart (Admin)
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /admin/dashboard/activity-chart [get]
func (h *DashboardHandler) GetActivityChart(c fiber.Ctx) error {
	res, err := h.svc.GetActivityChart(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "activity chart retrieved", res)
}

// GetPopulationMetrics handles GET /api/v1/staff/dashboard/population-metrics
// @Summary      Get population health metrics (Staff)
// @Description  Returns aggregated food intake, physical activity, and medication adherence for assigned patients. Accepts ?range=7,30,90
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Param        range query int false "Day range (7, 30, 90)" default(7)
// @Success      200  {object}  map[string]any
// @Router       /staff/dashboard/population-metrics [get]
func (h *DashboardHandler) GetPopulationMetrics(c fiber.Ctx) error {
	rangeStr := c.Query("range", "7")
	rangeDays, err := strconv.Atoi(rangeStr)
	if err != nil || rangeDays < 1 || rangeDays > 365 {
		rangeDays = 7
	}

	res, err := h.svc.GetPopulationMetrics(c.Context(), "", rangeDays)
	if err != nil {
		return err
	}
	return response.Success(c, "population metrics retrieved", res)
}

// GetPatientTrends handles GET /api/v1/staff/dashboard/patient-trends
// @Summary      Get patients with worsening health trends (Staff)
// @Description  Identifies patients whose average blood sugar is rising based on historical records within the specified range. Accepts ?range=7,30,90
// @Tags         dashboard
// @Security     BearerAuth
// @Produce      json
// @Param        range query int false "Day range (7, 30, 90)" default(7)
// @Success      200  {object}  map[string]any
// @Router       /staff/dashboard/patient-trends [get]
func (h *DashboardHandler) GetPatientTrends(c fiber.Ctx) error {
	rangeStr := c.Query("range", "7")
	rangeDays, err := strconv.Atoi(rangeStr)
	if err != nil || rangeDays < 1 || rangeDays > 365 {
		rangeDays = 7
	}

	res, err := h.svc.GetPatientTrends(c.Context(), "", rangeDays)
	if err != nil {
		return err
	}
	return response.Success(c, "patient trends retrieved", res)
}
