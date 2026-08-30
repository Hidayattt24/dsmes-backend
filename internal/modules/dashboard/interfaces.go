package dashboard

import "context"

type DashboardRepository interface {
	GetAdminStats(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffStats(ctx context.Context, healthFacility string) (*StaffDashboardResponse, error)
	GetTopArticles(ctx context.Context) ([]TopArticleResponse, error)
	GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error)
	GetPopulationMetrics(ctx context.Context, healthFacility string, rangeDays int) (*PopulationMetricsResponse, error)
	GetPatientTrends(ctx context.Context, healthFacility string, rangeDays int) ([]TrendPatient, error)
}

type DashboardService interface {
	GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffDashboard(ctx context.Context, staffID string) (*StaffDashboardResponse, error)
	GetTopArticles(ctx context.Context) ([]TopArticleResponse, error)
	GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error)
	GetPopulationMetrics(ctx context.Context, staffID string, rangeDays int) (*PopulationMetricsResponse, error)
	GetPatientTrends(ctx context.Context, staffID string, rangeDays int) ([]TrendPatient, error)
}
