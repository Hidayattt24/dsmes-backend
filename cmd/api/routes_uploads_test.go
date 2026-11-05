package main

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestUploadsStaticHandlerSupportsByteRanges(t *testing.T) {
	app := fiber.New()
	app.Use("/uploads", uploadsStaticHandler("testdata"))

	req := httptest.NewRequest("GET", "/uploads/sample.m4a", nil)
	req.Header.Set("Range", "bytes=0-3")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request upload range: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 206 {
		t.Fatalf("status = %d, want 206", resp.StatusCode)
	}
	if got := resp.Header.Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("Accept-Ranges = %q, want bytes", got)
	}
	if got := resp.Header.Get("Content-Range"); got != "bytes 0-3/11" {
		t.Fatalf("Content-Range = %q, want bytes 0-3/11", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got := string(body); got != "0123" {
		t.Fatalf("body = %q, want 0123", got)
	}
}
