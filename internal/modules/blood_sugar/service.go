package blood_sugar

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type bloodSugarService struct {
	repo BloodSugarRepository
	log  *zap.Logger
}

func NewBloodSugarService(repo BloodSugarRepository, log *zap.Logger) BloodSugarService {
	return &bloodSugarService{repo: repo, log: log}
}

func (s *bloodSugarService) LogBloodSugar(ctx context.Context, patientID string, req LogBloodSugarRequest) (*BloodSugarResponse, error) {
	measuredAt, err := time.Parse(time.RFC3339, req.MeasuredAt)
	if err != nil {
		return nil, errs.NewBadRequest("invalid measured_at format (RFC3339 required)", err)
	}

	status := domain.CalculateGlucoseStatus(req.GlucoseValue, req.MeasurementTimeType)

	log := &domain.BloodSugarLog{
		PatientID:           patientID,
		GlucoseValue:        req.GlucoseValue,
		MeasurementTimeType: req.MeasurementTimeType,
		MeasuredAt:          measuredAt,
		Status:              status,
	}

	if err = s.repo.Create(ctx, log); err != nil {
		return nil, err
	}

	res := ToBloodSugarResponse(log)
	return &res, nil
}

func (s *bloodSugarService) GetPatientHistory(ctx context.Context, patientID string, page, limit int) ([]BloodSugarResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAllByPatientID(ctx, patientID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]BloodSugarResponse, len(items))
	for i := range items {
		resp[i] = ToBloodSugarResponse(&items[i])
	}
	return resp, total, nil
}

func (s *bloodSugarService) GetPuskesmasDashboard(ctx context.Context, puskesmasID string) (*GlucoseDistributionResponse, error) {
	return s.repo.GetDistributionForPuskesmas(ctx, puskesmasID)
}
