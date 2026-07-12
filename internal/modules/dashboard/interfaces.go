package dashboard

import "context"

type DashboardRepository interface {
	GetAdminStats(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffStats(ctx context.Context, puskesmasID string) (*StaffDashboardResponse, error)
}

type DashboardService interface {
	GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffDashboard(ctx context.Context, puskesmasID string) (*StaffDashboardResponse, error)
}
