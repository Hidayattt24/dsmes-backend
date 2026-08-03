package education

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/middleware"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

type EducationHandler struct {
	svc EducationService
	log *zap.Logger
}

func NewEducationHandler(svc EducationService, log *zap.Logger) *EducationHandler {
	return &EducationHandler{svc: svc, log: log}
}

// ListCategories handles GET /api/v1/education/categories
// @Summary      Get list of categories
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /education/categories [get]
func (h *EducationHandler) ListCategories(c fiber.Ctx) error {
	items, err := h.svc.ListCategories(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "categories retrieved", items)
}

// ListPublished handles GET /api/v1/education/articles
// @Summary      Get list of published articles
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        category_id  query  string  false  "Filter by category"
// @Param        page         query  int     false  "Page number"
// @Param        limit        query  int     false  "Limit"
// @Success      200  {object}  map[string]any
// @Router       /education/articles [get]
func (h *EducationHandler) ListPublished(c fiber.Ctx) error {
	categoryID := c.Query("category_id")
	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 10
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	claims := middleware.ClaimsFromContext(c)
	var patientID *string
	if claims != nil && claims.Role == "user" {
		patientID = &claims.UserID
	}

	status := domain.StatusPublikasi
	items, total, err := h.svc.ListArticles(c.Context(), patientID, categoryID, &status, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "published articles retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetByID handles GET /api/v1/education/articles/:id
// @Summary      Get article details
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Article ID"
// @Success      200  {object}  map[string]any
// @Router       /education/articles/{id} [get]
func (h *EducationHandler) GetByID(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	var patientID *string
	if claims != nil && claims.Role == "user" {
		patientID = &claims.UserID
	}

	id := c.Params("id")
	res, err := h.svc.GetArticle(c.Context(), id, patientID)
	if err != nil {
		return err
	}
	return response.Success(c, "article details retrieved", res)
}

// Complete handles POST /api/v1/patient/education/:id/complete
// @Summary      Complete reading article
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Article ID"
// @Success      200  {object}  map[string]any
// @Router       /patient/education/{id}/complete [post]
func (h *EducationHandler) Complete(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	if err := h.svc.CompleteArticle(c.Context(), claims.UserID, id); err != nil {
		return err
	}
	return response.Success(c, "article completed", nil)
}

// Save handles POST /api/v1/patient/education/:id/save
// @Summary      Bookmark/save article
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Article ID"
// @Success      200  {object}  map[string]any
// @Router       /patient/education/{id}/save [post]
func (h *EducationHandler) Save(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	err := h.svc.SaveArticle(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}
	return response.Success(c, "article saved", map[string]bool{"saved": true})
}

// Unsave handles DELETE /api/v1/patient/education/:id/save
// @Summary      Unbookmark/unsave article
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Article ID"
// @Success      200  {object}  map[string]any
// @Router       /patient/education/{id}/save [delete]
func (h *EducationHandler) Unsave(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	err := h.svc.UnsaveArticle(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}
	return response.Success(c, "article unsaved", map[string]bool{"saved": false})
}

// ListSaved handles GET /api/v1/patient/education/saved
// @Summary      Get list of bookmarked articles
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /patient/education/saved [get]
func (h *EducationHandler) ListSaved(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	items, err := h.svc.ListSavedArticles(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, "bookmarked articles retrieved", items)
}

// ListAdmin handles GET /api/v1/admin/education/articles
// @Summary      List all articles (Admin)
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        category_id  query  string  false  "Filter by category"
// @Param        status       query  string  false  "Filter by status (draft/publikasi)"
// @Param        page         query  int     false  "Page number"
// @Param        limit        query  int     false  "Limit"
// @Success      200  {object}  map[string]any
// @Router       /admin/education/articles [get]
func (h *EducationHandler) ListAdmin(c fiber.Ctx) error {
	categoryID := c.Query("category_id")
	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil {
			page = p
		}
	}
	limit := 10
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			limit = l
		}
	}

	var s *domain.ArticleStatus
	statusStr := c.Query("status")
	if statusStr != "" {
		val := domain.ArticleStatus(statusStr)
		s = &val
	}

	items, total, err := h.svc.ListArticles(c.Context(), nil, categoryID, s, page, limit)
	if err != nil {
		return err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return response.SuccessWithMeta(c, "articles retrieved", items, &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Create handles POST /api/v1/admin/education/articles
// @Summary      Create article (Admin)
// @Tags         education
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateArticleRequest  true  "Create article payload"
// @Success      201  {object}  map[string]any
// @Router       /admin/education/articles [post]
func (h *EducationHandler) Create(c fiber.Ctx) error {
	claims := middleware.ClaimsFromContext(c)
	if claims == nil {
		return fiber.ErrUnauthorized
	}

	var req CreateArticleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.CreateArticle(c.Context(), claims.UserID, req)
	if err != nil {
		return err
	}
	return response.Created(c, "article created", res)
}

// Update handles PUT /api/v1/admin/education/articles/:id
// @Summary      Update article (Admin)
// @Tags         education
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  string                true  "Article ID"
// @Param        body  body  CreateArticleRequest  true  "Update article payload"
// @Success      200  {object}  map[string]any
// @Router       /admin/education/articles/{id} [put]
func (h *EducationHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req CreateArticleRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.UpdateArticle(c.Context(), id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "article updated", res)
}

// Publish handles PATCH /api/v1/admin/education/articles/:id/publish
// @Summary      Publish article (Admin)
// @Tags         education
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Article ID"
// @Success      200  {object}  map[string]any
// @Router       /admin/education/articles/{id}/publish [patch]
func (h *EducationHandler) Publish(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.PublishArticle(c.Context(), id); err != nil {
		return err
	}
	return response.Success(c, "article published successfully", nil)
}

// Delete handles DELETE /api/v1/admin/education/articles/:id
func (h *EducationHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteArticle(c.Context(), id); err != nil {
		return err
	}
	return response.NoContent(c)
}

// GetStats handles GET /api/v1/admin/education/stats
func (h *EducationHandler) GetStats(c fiber.Ctx) error {
	res, err := h.svc.GetStats(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, "education statistics retrieved", res)
}

func (h *EducationHandler) SubmitReview(c fiber.Ctx) error {
	id := c.Params("id")
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	var req CreateReviewRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}

	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.SubmitReview(c.Context(), claims.UserID, id, req)
	if err != nil {
		return err
	}
	return response.Success(c, "review submitted successfully", res)
}

func (h *EducationHandler) GetReview(c fiber.Ctx) error {
	id := c.Params("id")
	claims := middleware.ClaimsFromContext(c)
	if claims == nil || claims.UserID == "" {
		return errs.NewUnauthorized("unauthorized")
	}

	res, err := h.svc.GetPatientReview(c.Context(), claims.UserID, id)
	if err != nil {
		return err
	}
	return response.Success(c, "review retrieved", res)
}

func (h *EducationHandler) GetRatingSummary(c fiber.Ctx) error {
	id := c.Params("id")
	claims := middleware.ClaimsFromContext(c)
	var patientID *string
	if claims != nil && claims.Role == "user" {
		patientID = &claims.UserID
	}

	res, err := h.svc.GetRatingSummary(c.Context(), id, patientID)
	if err != nil {
		return err
	}
	return response.Success(c, "rating summary retrieved", res)
}

func (h *EducationHandler) GetAdminReviews(c fiber.Ctx) error {
	id := c.Params("id")
	res, err := h.svc.GetAdminReviews(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, "admin article reviews retrieved", res)
}

