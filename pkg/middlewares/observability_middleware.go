package middlewares

import (
	"log"
	"strings"
	"time"

	"hifzhun-api/pkg/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ObservabilityMiddleware logs HTTP latency, tags slow requests, and logs DB pool snapshots on delay
func ObservabilityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Resolve or generate a safe request ID
		reqID := c.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()[:8]
		}
		c.Locals("request_id", reqID)
		c.Set("X-Request-ID", reqID)

		// 2. Measure execution time
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		durationMs := duration.Milliseconds()

		statusCode := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		// Skip high-frequency Swagger / static upload logs from noisy log pollution
		if strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/uploads") {
			return err
		}

		timestamp := start.UTC().Format(time.RFC3339)

		// Standard HTTP latency log (grep: [HTTP])
		log.Printf("[HTTP] timestamp=\"%s\" request_id=%s method=%s path=%s status=%d duration_ms=%d",
			timestamp, reqID, method, path, statusCode, durationMs)

		// Slow request classification & DB pool snapshot
		if durationMs >= 1000 {
			var tag string
			switch {
			case durationMs >= 10000:
				tag = "[TIMEOUT_LEVEL]"
			case durationMs >= 8000:
				tag = "[TIMEOUT_RISK]"
			default:
				tag = "[SLOW_REQUEST]"
			}

			log.Printf("%s request_id=%s method=%s path=%s status=%d duration_ms=%d",
				tag, reqID, method, path, statusCode, durationMs)

			// Capture runtime DB pool stats without querying Postgres (in-memory Go sql.DB.Stats)
			if stats, ok := config.GetDBPoolStats(); ok {
				log.Printf("[DB_POOL] request_id=%s open=%d in_use=%d idle=%d max_open=%d wait_count=%d wait_duration_ms=%d max_idle_closed=%d max_idle_time_closed=%d max_lifetime_closed=%d",
					reqID,
					stats.OpenConnections,
					stats.InUse,
					stats.Idle,
					stats.MaxOpenConnections,
					stats.WaitCount,
					stats.WaitDurationMs,
					stats.MaxIdleClosed,
					stats.MaxIdleTimeClosed,
					stats.MaxLifetimeClosed,
				)
			}
		}

		return err
	}
}
