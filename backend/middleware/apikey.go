package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIKeyAuth middleware validates API keys from the X-API-Key header
func APIKeyAuth(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "X-API-Key header required",
			})
		}

		// Hash the provided key and look it up
		hash := sha256.Sum256([]byte(apiKey))
		keyHash := fmt.Sprintf("%x", hash)

		var projectID string
		var isActive bool
		err := db.QueryRow(context.Background(),
			`SELECT project_id, is_active FROM api_keys WHERE key_hash = $1`,
			keyHash,
		).Scan(&projectID, &isActive)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		if !isActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "API key is inactive",
			})
		}

		// Update last used timestamp
		go func() {
			_, _ = db.Exec(context.Background(),
				`UPDATE api_keys SET last_used_at = $1 WHERE key_hash = $2`,
				time.Now(), keyHash,
			)
		}()

		c.Locals("project_id", projectID)
		return c.Next()
	}
}
