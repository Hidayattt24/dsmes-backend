package dashboard

import "context"

type DashboardRepository interface {
	GetAdminStats(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffStats(ctx context.Context, staffID string) (*StaffDashboardResponse, error)
	GetTopArticles(ctx context.Context) ([]TopArticleResponse, error)
	GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error)
}

type DashboardService interface {
	GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffDashboard(ctx context.Context, staffID string) (*StaffDashboardResponse, error)
	GetTopArticles(ctx context.Context) ([]TopArticleResponse, error)
	GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error)
}
