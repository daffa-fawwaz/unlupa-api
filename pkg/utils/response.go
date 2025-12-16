package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Meta struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
	Total   int `json:"total,omitempty"`
}

type SuccessResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
}

type FieldError struct {
	Field    string   `json:"field"`
	Messages []string `json:"messages"`
	Message  string   `json:"message"`
}

type ErrorResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Error      string       `json:"error"`
	StatusCode int          `json:"statusCode"`
	Timestamp  string       `json:"timestamp"`
	Path       string       `json:"path"`
	Errors     []FieldError `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data interface{}, meta *Meta) error {
	resp := SuccessResponse{
		Status:    status,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      c.OriginalURL(),
	}
	return c.Status(status).JSON(resp)
}

func Error(c *fiber.Ctx, status int, message string, errType string, fieldErrors []FieldError) error {
	resp := ErrorResponse{
		Success:    false,
		Message:    message,
		Error:      errType,
		StatusCode: status,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Path:       c.OriginalURL(),
		Errors:     fieldErrors,
	}
	return c.Status(status).JSON(resp)
}
