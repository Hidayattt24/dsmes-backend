package _template

import (
	"github.com/dsmes/dsmes-backend/internal/domain"
)

// Template is the GORM entity for this module's database table.
// Rules:
//   - Always embed domain.BaseModel (UUID PK + timestamps + soft-delete)
//   - Use gorm struct tags for constraints and indexes
//   - Never expose this struct in API responses — map to TemplateResponse DTO
//
// TableName overrides the default GORM table name convention.
// GORM default would be "templates"; be explicit to prevent surprises.
type Template struct {
	domain.BaseModel `gorm:"embedded"`

	Name string `gorm:"type:varchar(150);not null"`
}

// TableName sets the database table name for this model.
func (Template) TableName() string {
	return "templates"
}

// ToResponse converts the domain model to the API response DTO.
// This is the mapping layer that prevents direct model exposure.
func (t *Template) ToResponse() TemplateResponse {
	return TemplateResponse{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
