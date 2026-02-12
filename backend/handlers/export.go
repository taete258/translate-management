package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"translate-management/cache"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vmihailenco/msgpack/v5"
)

type ExportHandler struct {
	DB    *pgxpool.Pool
	Cache *cache.RedisClient
}

func NewExportHandler(db *pgxpool.Pool, rdb *cache.RedisClient) *ExportHandler {
	return &ExportHandler{DB: db, Cache: rdb}
}

// Export returns translations for a project/language in JSON or MessagePack format
func (h *ExportHandler) Export(c *fiber.Ctx) error {
	slug := c.Params("slug")
	langCode := c.Params("langCode")
	format := c.Query("format", "json")

	if format != "json" && format != "msgpack" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format must be 'json' or 'msgpack'"})
	}

	// Check cache first
	cacheKey := cache.CacheKey(slug, langCode, format)
	cached, err := h.Cache.Get(context.Background(), cacheKey)
	if err == nil && cached != nil {
		if format == "msgpack" {
			c.Set("Content-Type", "application/x-msgpack")
		} else {
			c.Set("Content-Type", "application/json")
		}
		c.Set("X-Cache", "HIT")
		return c.Send(cached)
	}

	// Get project ID from slug
	// Get project ID from slug
	var projectID string
	err = h.DB.QueryRow(context.Background(),
		`SELECT id FROM projects WHERE slug = $1`, slug,
	).Scan(&projectID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	// Verify API key belongs to this project
	authProjectID := c.Locals("project_id")
	if authProjectID != nil && authProjectID.(string) != projectID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key does not belong to this project"})
	}

	// Get language ID
	var languageID string
	err = h.DB.QueryRow(context.Background(),
		`SELECT id FROM languages WHERE project_id = $1 AND code = $2`, projectID, langCode,
	).Scan(&languageID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found"})
	}

	// Get all translations for this language
	rows, err := h.DB.Query(context.Background(),
		`SELECT tk.key, t.value
		 FROM translation_keys tk
		 LEFT JOIN translations t ON t.key_id = tk.id AND t.language_id = $2
		 WHERE tk.project_id = $1
		 ORDER BY tk.key`,
		projectID, languageID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch translations"})
	}
	defer rows.Close()

	flatMap := make(map[string]string)
	for rows.Next() {
		var key string
		var value *string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		if value != nil {
			flatMap[key] = *value
		} else {
			flatMap[key] = ""
		}
	}

	// Build nested structure from dot-notation keys
	nested := buildNestedMap(flatMap)

	var data []byte
	if format == "msgpack" {
		data, err = msgpack.Marshal(nested)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode MessagePack"})
		}
		c.Set("Content-Type", "application/x-msgpack")
	} else {
		data, err = json.MarshalIndent(nested, "", "  ")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode JSON"})
		}
		c.Set("Content-Type", "application/json")
	}

	// Cache the result for 1 hour
	_ = h.Cache.Set(context.Background(), cacheKey, data, 1*time.Hour)

	c.Set("X-Cache", "MISS")
	return c.Send(data)
}

// buildNestedMap converts flat dot-notation keys to nested maps
// e.g. {"home.hero.title": "Hello"} -> {"home": {"hero": {"title": "Hello"}}}
func buildNestedMap(flatMap map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range flatMap {
		parts := strings.Split(key, ".")
		current := result

		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
			} else {
				if _, ok := current[part]; !ok {
					current[part] = make(map[string]interface{})
				}
				if next, ok := current[part].(map[string]interface{}); ok {
					current = next
				} else {
					// Key conflict: a value exists where we need a map
					newMap := make(map[string]interface{})
					current[part] = newMap
					current = newMap
				}
			}
		}
	}

	return result
}
