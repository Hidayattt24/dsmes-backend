// Package _template is a documentation skeleton that shows the exact
// Clean Architecture layer structure every business module must follow.
//
// HOW TO USE THIS TEMPLATE
// ────────────────────────
// 1. Copy this directory: cp -r internal/modules/_template internal/modules/<your_module>
// 2. Rename the package from "_template" to your module name (e.g. "auth", "patient")
// 3. Replace every "Template" prefix with your entity name
// 4. Delete this comment block
//
// LAYER RULES (strictly enforced)
// ──────────────────────────────
//   Handler    → HTTP only. Parses request, calls Service, writes response.
//   Service    → Business logic only. Depends on Repository interface.
//   Repository → Data access only. Depends on *gorm.DB.
//   DTO        → Input/Output shapes. No DB tags. No domain logic.
//   Routes     → Mounts handler methods on Fiber route groups.
//
// DEPENDENCY DIRECTION (Clean Architecture inward rule)
//   Handler → Service interface → Repository interface
//   Concrete implementations flow outward via constructor injection.
//
// DO NOT:
//   - Put business logic in Handler
//   - Import gorm or sql in Handler or Service
//   - Expose domain.BaseModel or GORM models in API responses
//   - Use global variables for dependencies
package _template
