package education_progress

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type EducationProgressHandler struct {
	svc EducationProgressService
	log *zap.Logger
}

func NewEducationProgressHandler(svc EducationProgressService, log *zap.Logger) *EducationProgressHandler {
	return &EducationProgressHandler{svc: svc, log: log}
}

// GetArticleProgress handles GET /api/v1/admin/education/:id/progress
// Returns all patients' progress for a given education article.
func (h *EducationProgressHandler) GetArticleProgress(c fiber.Ctx) error {
	articleID := c.Params("id")
	items, err := h.svc.GetArticleProgress(c.Context(), articleID)
	if err != nil {
		return err
	}
	if items == nil {
		items = []PatientProgressItem{}
	}
	return response.Success(c, "education progress retrieved", items)
}

// GetArticleAnalytics handles GET /api/v1/admin/education/:id/progress/analytics
// Returns summary statistics for a given education article.
func (h *EducationProgressHandler) GetArticleAnalytics(c fiber.Ctx) error {
	articleID := c.Params("id")
	stats, err := h.svc.GetArticleAnalytics(c.Context(), articleID)
	if err != nil {
		return err
	}
	return response.Success(c, "education progress analytics retrieved", stats)
}

// MarkArticleRead handles POST /api/v1/patient/education/:id/read-article
// Marks the article as read for the authenticated patient.
func (h *EducationProgressHandler) MarkArticleRead(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req MarkArticleReadRequest
	_ = c.Bind().Body(&req)

	articleID := c.Params("id")
	if err := h.svc.MarkArticleRead(c.Context(), claims.UserID, articleID, req.ReadingDuration, req.LastScrollPosition); err != nil {
		return err
	}
	return response.Success(c, "article marked as read", nil)
}

// MarkVideoWatched handles POST /api/v1/patient/education/:id/watch-video
// Marks the video as watched for the authenticated patient.
func (h *EducationProgressHandler) MarkVideoWatched(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req MarkVideoWatchedRequest
	_ = c.Bind().Body(&req)

	articleID := c.Params("id")
	if err := h.svc.MarkVideoWatched(c.Context(), claims.UserID, articleID, req.WatchDuration, req.VideoLastTimestamp); err != nil {
		return err
	}
	return response.Success(c, "video marked as watched", nil)
}

// GetPatientProgress handles GET /api/v1/patient/education/:id/progress
// Returns the authenticated patient's progress for an education article.
func (h *EducationProgressHandler) GetPatientProgress(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	articleID := c.Params("id")
	item, err := h.svc.GetPatientProgress(c.Context(), claims.UserID, articleID)
	if err != nil {
		return err
	}
	return response.Success(c, "patient progress retrieved", item)
}

// GetPatientEducationActivities handles GET /api/v1/admin/patients/:id/education-activities
// Returns all education activities for a specific patient.
func (h *EducationProgressHandler) GetPatientEducationActivities(c fiber.Ctx) error {
	patientID := c.Params("id")
	result, err := h.svc.GetPatientEducationActivities(c.Context(), patientID)
	if err != nil {
		return err
	}
	return response.Success(c, "patient education activities retrieved", result)
}

// MarkArticleReadAdmin handles POST /api/v1/admin/education/:id/progress/read-article
// Admin marks article as read for a specific patient.
func (h *EducationProgressHandler) MarkArticleReadAdmin(c fiber.Ctx) error {
	articleID := c.Params("id")
	var req MarkActionRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}
	// Admin force-marks with duration=0 and scroll=100 (fully read)
	if err := h.svc.MarkArticleRead(c.Context(), req.PatientID, articleID, 0, 100); err != nil {
		return err
	}
	return response.Success(c, "article marked as read", nil)
}

// MarkVideoWatchedAdmin handles POST /api/v1/admin/education/:id/progress/watch-video
// Admin marks video as watched for a specific patient.
func (h *EducationProgressHandler) MarkVideoWatchedAdmin(c fiber.Ctx) error {
	articleID := c.Params("id")
	var req MarkActionRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}
	// Admin force-marks with duration=0 and timestamp=0
	if err := h.svc.MarkVideoWatched(c.Context(), req.PatientID, articleID, 0, 0); err != nil {
		return err
	}
	return response.Success(c, "video marked as watched", nil)
}
