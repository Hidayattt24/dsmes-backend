// Package response provides a standardised JSON response wrapper.
//
// Architectural decision: every API response (success OR error) MUST go through
// this package. This guarantees a consistent JSON envelope across all endpoints,
// which is required by the api-design skill rule ("Response DTO" / "Never expose
// internal database model").
//
// Response envelope:
//
//	{
//	  "success": true,
//	  "message": "data retrieved",
//	  "data": { ... },         // omitted on error
//	  "errors": [ ... ],       // omitted on success
//	  "meta": { ... }          // optional pagination / additional metadata
//	}
package response

import (
	"github.com/gofiber/fiber/v3"
)

// Meta holds optional pagination or extra metadata attached to list responses.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// envelope is the canonical JSON structure for every API response.
type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Success sends a 200 OK response with the provided data and message.
func Success(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a 200 OK response with data and pagination metadata.
func SuccessWithMeta(c fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 Created response.
func Created(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response (empty body).
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Error sends a structured error response.
// status is the HTTP status code; message is exposed to the caller.
// errs (optional) can carry field-level validation errors.
func Error(c fiber.Ctx, status int, message string, errs ...any) error {
	env := envelope{
		Success: false,
		Message: message,
	}
	if len(errs) > 0 {
		env.Errors = errs[0]
	}
	return c.Status(status).JSON(env)
}

// ValidationError sends a 422 Unprocessable Entity with structured field errors.
// Used by the DTO validator to surface per-field error messages.
func ValidationError(c fiber.Ctx, fieldErrors any) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(envelope{
		Success: false,
		Message: "validation failed",
		Errors:  fieldErrors,
	})
}
