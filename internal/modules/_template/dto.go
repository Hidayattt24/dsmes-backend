package _template

// TemplateRequest is the DTO for creating/updating a template entity.
// Rules:
//   - Use json tags (not gorm tags) — DTOs are never persisted directly
//   - Use validate tags for request validation
//   - Never embed domain.BaseModel or expose DB column names
//
// Example usage in handler:
//
//	var req TemplateRequest
//	if err := c.Bind().Body(&req); err != nil {
//	    return response.Error(c, fiber.StatusBadRequest, "invalid request body")
//	}
//	if errs := validator.Validate(&req); errs != nil {
//	    return response.ValidationError(c, errs)
//	}
type TemplateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=150"`
}

// TemplateResponse is the DTO returned to API clients.
// Rules:
//   - Never expose password_hash, internal IDs, or audit fields the client doesn't need
//   - Use json tags with snake_case field names
//   - Add omitempty only for genuinely optional fields
type TemplateResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// TemplateListResponse wraps a slice of responses for list endpoints.
// The response package's SuccessWithMeta() sends the pagination envelope.
type TemplateListResponse struct {
	Items []TemplateResponse `json:"items"`
}
