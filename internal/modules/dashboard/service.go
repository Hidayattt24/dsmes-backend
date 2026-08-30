package dashboard

import (
	"context"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/modules/facility"
	"github.com/dsmes/dsmes-backend/internal/modules/staff"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type dashboardService struct {
	repo         DashboardRepository
	staffRepo    staff.StaffRepository
	facilityRepo facility.FacilityRepository
	log          *zap.Logger
}

func NewDashboardService(repo DashboardRepository, staffRepo staff.StaffRepository, facilityRepo facility.FacilityRepository, log *zap.Logger) DashboardService {
	return &dashboardService{repo: repo, staffRepo: staffRepo, facilityRepo: facilityRepo, log: log}
}

func (s *dashboardService) GetAdminDashboard(ctx context.Context) (*AdminDashboardResponse, error) {
	return s.repo.GetAdminStats(ctx)
}

// resolveFacilityName maps a staff account to its assigned Puskesmas name.
func (s *dashboardService) resolveFacilityName(ctx context.Context, staffID string) (string, error) {
	if staffID == "" {
		return "", errs.NewUnauthorized("staff not identified")
	}
	sa, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return "", err
	}
	if sa.HealthFacilityID == nil || *sa.HealthFacilityID == "" {
		return "", errs.NewForbidden("staff belum memiliki puskesmas")
	}
	f, err := s.facilityRepo.FindByID(ctx, *sa.HealthFacilityID)
	if err != nil {
		return "", err
	}
	return f.Name, nil
}

func (s *dashboardService) GetStaffDashboard(ctx context.Context, staffID string) (*StaffDashboardResponse, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetStaffStats(ctx, name)
}

func (s *dashboardService) GetTopArticles(ctx context.Context) ([]TopArticleResponse, error) {
	return s.repo.GetTopArticles(ctx)
}

func (s *dashboardService) GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error) {
	return s.repo.GetActivityChart(ctx)
}

func (s *dashboardService) GetPopulationMetrics(ctx context.Context, staffID string, rangeDays int) (*PopulationMetricsResponse, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPopulationMetrics(ctx, name, rangeDays)
}

func (s *dashboardService) GetPatientTrends(ctx context.Context, staffID string, rangeDays int) ([]TrendPatient, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetPatientTrends(ctx, name, rangeDays)
}
