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

	normType := domain.NormalizeMeasurementType(string(req.MeasurementTimeType))
	dob, _ := s.repo.GetPatientDOB(ctx, patientID)
	medRes := domain.CalculateBloodSugarMedicalResult(req.GlucoseValue, normType, dob)

	log := &domain.BloodSugarLog{
		PatientID:           patientID,
		GlucoseValue:        req.GlucoseValue,
		MeasurementTimeType: normType,
		MeasuredAt:          measuredAt,
		Status:              medRes.Classification,
		Severity:            medRes.Severity,
		ReferenceMin:        medRes.ReferenceMin,
		ReferenceMax:        medRes.ReferenceMax,
		ReferenceRangeText:  medRes.ReferenceRangeText,
		Recommendation:      medRes.Recommendation,
		ColorIndicator:      medRes.ColorIndicator,
	}

	if err = s.repo.Create(ctx, log); err != nil {
		return nil, err
	}

	res := ToBloodSugarResponse(log)
	return &res, nil
}

func (s *bloodSugarService) UpdateBloodSugar(ctx context.Context, patientID, id string, req LogBloodSugarRequest) (*BloodSugarResponse, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.PatientID != patientID {
		return nil, errs.NewForbidden("unauthorized access to blood sugar log")
	}

	measuredAt, err := time.Parse(time.RFC3339, req.MeasuredAt)
	if err != nil {
		return nil, errs.NewBadRequest("invalid measured_at format (RFC3339 required)", err)
	}

	normType := domain.NormalizeMeasurementType(string(req.MeasurementTimeType))
	dob, _ := s.repo.GetPatientDOB(ctx, patientID)
	medRes := domain.CalculateBloodSugarMedicalResult(req.GlucoseValue, normType, dob)

	existing.GlucoseValue = req.GlucoseValue
	existing.MeasurementTimeType = normType
	existing.MeasuredAt = measuredAt
	existing.Status = medRes.Classification
	existing.Severity = medRes.Severity
	existing.ReferenceMin = medRes.ReferenceMin
	existing.ReferenceMax = medRes.ReferenceMax
	existing.ReferenceRangeText = medRes.ReferenceRangeText
	existing.Recommendation = medRes.Recommendation
	existing.ColorIndicator = medRes.ColorIndicator

	if err = s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	res := ToBloodSugarResponse(existing)
	return &res, nil
}

func (s *bloodSugarService) DeleteBloodSugar(ctx context.Context, patientID, id string) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.PatientID != patientID {
		return errs.NewForbidden("unauthorized access to blood sugar log")
	}
	return s.repo.Delete(ctx, id)
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

func (s *bloodSugarService) GetStaffDashboard(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error) {
	return s.repo.GetDistributionForStaff(ctx, staffID)
}
