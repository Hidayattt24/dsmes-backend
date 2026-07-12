package dashboard

import "context"

type DashboardRepository interface {
	GetAdminStats(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffStats(ctx context.Context, staffID string) (*StaffDashboardResponse, error)
}

type DashboardService interface {
	GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error)
	GetStaffDashboard(ctx context.Context, staffID string) (*StaffDashboardResponse, error)
}
