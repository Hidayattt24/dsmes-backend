package _template

import (
	"context"

	"go.uber.org/zap"
)

// templateService is the concrete implementation of TemplateService.
// It depends on TemplateRepository (the interface, not the struct).
type templateService struct {
	repo TemplateRepository
	log  *zap.Logger
}

// NewTemplateService creates a new service instance.
// Receives the repository interface — never *gorm.DB directly.
func NewTemplateService(repo TemplateRepository, log *zap.Logger) TemplateService {
	return &templateService{repo: repo, log: log}
}

// List retrieves a paginated list of templates and maps them to response DTOs.
func (s *templateService) List(ctx context.Context, page, limit int) ([]TemplateResponse, int64, error) {
	// Default pagination guards
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]TemplateResponse, 0, len(items))
	for i := range items {
		responses = append(responses, items[i].ToResponse())
	}

	return responses, total, nil
}

// GetByID retrieves a single template by UUID and maps it to a response DTO.
func (s *templateService) GetByID(ctx context.Context, id string) (*TemplateResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := item.ToResponse()
	return &res, nil
}

// Create validates business rules and persists a new template.
func (s *templateService) Create(ctx context.Context, req TemplateRequest) (*TemplateResponse, error) {
	// Place additional business-rule validation here.
	// Struct-tag validation has already run in the handler via validator.Validate().

	item := &Template{Name: req.Name}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	res := item.ToResponse()
	return &res, nil
}

// Update validates business rules and saves changes to an existing template.
func (s *templateService) Update(ctx context.Context, id string, req TemplateRequest) (*TemplateResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.Name = req.Name

	if err = s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	res := item.ToResponse()
	return &res, nil
}

// Delete removes a template by UUID (soft-delete via GORM).
func (s *templateService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
