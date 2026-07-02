package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppErrorFormatting(t *testing.T) {
	err := New(http.StatusBadRequest, "invalid request")
	if err.HTTPCode() != http.StatusBadRequest {
		t.Errorf("expected HTTP code 400, got: %d", err.HTTPCode())
	}
	if err.Error() != "[400] invalid request" {
		t.Errorf("expected error string '[400] invalid request', got: %s", err.Error())
	}
}

func TestAppErrorWrapping(t *testing.T) {
	cause := errors.New("db connection failure")
	err := NewInternal("internal error occurred", cause)

	if !errors.Is(err, cause) {
		t.Error("expected AppError to wrap and match inner error via errors.Is")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError to be extractable via errors.As")
	}

	if appErr.Code != http.StatusInternalServerError {
		t.Errorf("expected inner code to be 500, got: %d", appErr.Code)
	}
}

func TestSentinelHelperFunctions(t *testing.T) {
	notFoundErr := NewNotFound("user not found")
	conflictErr := NewConflict("email already registered")
	genericErr := errors.New("generic error")

	if !IsNotFound(notFoundErr) {
		t.Error("expected IsNotFound to return true for NotFound error")
	}
	if IsNotFound(genericErr) {
		t.Error("expected IsNotFound to return false for generic error")
	}

	if !IsConflict(conflictErr) {
		t.Error("expected IsConflict to return true for Conflict error")
	}
	if IsConflict(genericErr) {
		t.Error("expected IsConflict to return false for generic error")
	}
}
