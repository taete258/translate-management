package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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

	data, err := h.getOrGenerateData(slug, langCode, format, c)
	if err != nil {
		return err
	}

	if format == "msgpack" {
		c.Set("Content-Type", "application/x-msgpack")
	} else {
		c.Set("Content-Type", "application/json")
	}
	// X-Cache header is already set in getOrGenerateData if HIT
	// But if MISS, we need to set it? getOrGenerateData sets it to MISS on generation.
    
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


// GetVersion returns the version hash of the translations
func (h *ExportHandler) GetVersion(c *fiber.Ctx) error {
	slug := c.Params("slug")
	langCode := c.Params("langCode")
	format := c.Query("format", "json")

	if format != "json" && format != "msgpack" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format must be 'json' or 'msgpack'"})
	}

	// 1. Try to get from cache
	cacheKey := cache.CacheKey(slug, langCode, format)
	cached, err := h.Cache.Get(context.Background(), cacheKey)

	if err != nil || cached == nil {
		// Cache miss: Generate the export to populate cache
		// We can reuse the logic by calling a helper or just invoking Export implementation.
		// However, calling Export directly writes to response. We need the data.
		// So we'll refactor Export to separate data generation from response writing?
		// Or simpler: just duplicate the generation logic for now to avoid big refactor risk,
		// OR better: call the generation logic.
		// Let's refactor slightly to extract generation logic.
		
		// Wait, refactoring might be risky. Let's look at Export again.
		// It does: DB queries -> buildNestedMap -> Marshal -> Cache.Set -> Send.
		
		// To avoid duplication, let's extract the generation logic.
		data, err := h.generateExportData(slug, langCode, format, c)
		if err != nil {
			return err // generateExportData handles error responses
		}
		cached = data
	}

	// 2. Calculate Hash
	hash := calculateHash(cached)

	return c.JSON(fiber.Map{
		"version": hash,
	})
}

func (h *ExportHandler) generateExportData(slug, langCode, format string, c *fiber.Ctx) ([]byte, error) {
	// Re-implementing specific parts of Export for internal use
	// Note: This duplicates logic from Export. ideally we refactor Export to use this.
	// But to minimize changes to Export for now, I will implement this
	// ensuring it populates the cache.

	// Actually, let's just Refactor Export to be safe and clean.
	return h.getOrGenerateData(slug, langCode, format, c)
}

// getOrGenerateData handles the core logic: check cache, if miss -> generate & set cache
func (h *ExportHandler) getOrGenerateData(slug, langCode, format string, c *fiber.Ctx) ([]byte, error) {
	cacheKey := cache.CacheKey(slug, langCode, format)
	cached, err := h.Cache.Get(context.Background(), cacheKey)
	if err == nil && cached != nil {
		c.Set("X-Cache", "HIT")
		return cached, nil
	}

	// --- GENERATION LOGIC START ---
	var projectID string
	err = h.DB.QueryRow(context.Background(),
		`SELECT id FROM projects WHERE slug = $1`, slug,
	).Scan(&projectID)

	if err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
		return nil, err
	}

	// Verify API key belongs to this project
	authProjectID := c.Locals("project_id")
	if authProjectID != nil && authProjectID.(string) != projectID {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "API key does not belong to this project"})
		return nil, fiber.NewError(fiber.StatusForbidden, "API key does not belong to this project")
	}

	var languageID string
	err = h.DB.QueryRow(context.Background(),
		`SELECT id FROM languages WHERE project_id = $1 AND code = $2`, projectID, langCode,
	).Scan(&languageID)

	if err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found"})
		return nil, err
	}

	rows, err := h.DB.Query(context.Background(),
		`SELECT tk.key, t.value
		 FROM translation_keys tk
		 LEFT JOIN translations t ON t.key_id = tk.id AND t.language_id = $2
		 WHERE tk.project_id = $1
		 ORDER BY tk.key`,
		projectID, languageID,
	)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch translations"})
		return nil, err
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

	nested := buildNestedMap(flatMap)

	var data []byte
	if format == "msgpack" {
		data, err = msgpack.Marshal(nested)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode MessagePack"})
			return nil, err
		}
	} else {
		data, err = json.MarshalIndent(nested, "", "  ")
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode JSON"})
			return nil, err
		}
	}

	// Cache the result for 1 hour
	_ = h.Cache.Set(context.Background(), cacheKey, data, 1*time.Hour)
	c.Set("X-Cache", "MISS")
	
	return data, nil
}

func calculateHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
