package dashboard

import (
	"context"

	"go.uber.org/zap"
)

type dashboardService struct {
	repo DashboardRepository
	log  *zap.Logger
}

func NewDashboardService(repo DashboardRepository, log *zap.Logger) DashboardService {
	return &dashboardService{repo: repo, log: log}
}

func (s *dashboardService) GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error) {
	return s.repo.GetAdminStats(ctx)
}

func (s *dashboardService) GetStaffDashboard(ctx context.Context, staffID string) (*StaffDashboardResponse, error) {
	return s.repo.GetStaffStats(ctx, staffID)
}
