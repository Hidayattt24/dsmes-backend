package response

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestResponseSuccess(t *testing.T) {
	app := fiber.New()

	app.Get("/test-success", func(c fiber.Ctx) error {
		return Success(c, "operation successful", map[string]string{"id": "123"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test-success", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var res envelope
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !res.Success {
		t.Error("expected success field to be true")
	}

	if res.Message != "operation successful" {
		t.Errorf("expected message 'operation successful', got: %s", res.Message)
	}

	dataMap, ok := res.Data.(map[string]any)
	if !ok || dataMap["id"] != "123" {
		t.Errorf("expected data to contain id '123', got: %v", res.Data)
	}
}

func TestResponseError(t *testing.T) {
	app := fiber.New()

	app.Get("/test-error", func(c fiber.Ctx) error {
		return Error(c, http.StatusBadRequest, "bad request parameter", []string{"field_invalid"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test-error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 Bad Request, got: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var res envelope
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if res.Success {
		t.Error("expected success field to be false")
	}

	if res.Message != "bad request parameter" {
		t.Errorf("expected message 'bad request parameter', got: %s", res.Message)
	}

	errList, ok := res.Errors.([]any)
	if !ok || len(errList) != 1 || errList[0] != "field_invalid" {
		t.Errorf("expected errors list to contain 'field_invalid', got: %v", res.Errors)
	}
}
