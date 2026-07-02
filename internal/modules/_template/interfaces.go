package _template

import (
	"context"
)

// TemplateRepository defines the data access contract for the Template entity.
//
// Rules:
//   - This is an INTERFACE — the service layer depends on this abstraction
//   - The concrete implementation lives in repository.go
//   - All methods accept a context.Context for cancellation and tracing
//   - Use errs.NewNotFound / errs.NewInternal to wrap DB errors
//   - Never return raw gorm.ErrRecordNotFound or sql.ErrNoRows to callers
//
// Why an interface?
//   The service never imports gorm directly. This makes the service fully
//   testable: swap in a mock repository in unit tests without a real database.
type TemplateRepository interface {
	// FindAll returns a paginated list of templates.
	FindAll(ctx context.Context, page, limit int) ([]Template, int64, error)

	// FindByID returns a single template by its UUID primary key.
	// Returns errs.NewNotFound if no record exists.
	FindByID(ctx context.Context, id string) (*Template, error)

	// Create persists a new template and returns the saved entity.
	Create(ctx context.Context, t *Template) error

	// Update saves changes to an existing template.
	// Returns errs.NewNotFound if the record does not exist.
	Update(ctx context.Context, t *Template) error

	// Delete soft-deletes a template by its UUID.
	// Returns errs.NewNotFound if the record does not exist.
	Delete(ctx context.Context, id string) error
}

// TemplateService defines the business logic contract for the Template module.
//
// Rules:
//   - This is an INTERFACE — handlers depend on this abstraction
//   - The concrete implementation lives in service.go
//   - Input comes as Request DTOs; output goes as Response DTOs
//   - Business validation (beyond struct-tag validation) lives here
type TemplateService interface {
	// List returns a paginated list of template response DTOs.
	List(ctx context.Context, page, limit int) ([]TemplateResponse, int64, error)

	// GetByID returns a single template response by UUID.
	GetByID(ctx context.Context, id string) (*TemplateResponse, error)

	// Create validates the request and creates a new template.
	Create(ctx context.Context, req TemplateRequest) (*TemplateResponse, error)

	// Update validates the request and updates an existing template.
	Update(ctx context.Context, id string, req TemplateRequest) (*TemplateResponse, error)

	// Delete removes a template by UUID.
	Delete(ctx context.Context, id string) error
}
