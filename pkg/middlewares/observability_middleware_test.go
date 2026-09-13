package middlewares_test

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func TestObservabilityMiddleware_NormalRequest(t *testing.T) {
	app := fiber.New()
	app.Use(middlewares.ObservabilityMiddleware())

	app.Get("/test-normal", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test-normal", nil)
	resp, err := app.Test(req, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Errorf("expected X-Request-ID header to be present")
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got '%s'", string(body))
	}
}

func TestObservabilityMiddleware_SlowRequest(t *testing.T) {
	app := fiber.New()
	app.Use(middlewares.ObservabilityMiddleware())

	app.Get("/test-slow", func(c *fiber.Ctx) error {
		time.Sleep(1050 * time.Millisecond)
		return c.SendString("slow response")
	})

	req := httptest.NewRequest("GET", "/test-slow", nil)
	req.Header.Set("X-Request-ID", "custom-req-id")
	resp, err := app.Test(req, 2000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	reqID := resp.Header.Get("X-Request-ID")
	if reqID != "custom-req-id" {
		t.Errorf("expected X-Request-ID 'custom-req-id', got '%s'", reqID)
	}
}
