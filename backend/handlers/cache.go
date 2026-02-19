package handlers

import (
	"context"
	"encoding/json"
	"time"

	"translate-management/cache"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vmihailenco/msgpack/v5"
)

type CacheHandler struct {
	DB    *pgxpool.Pool
	Cache *cache.RedisClient
}

func NewCacheHandler(db *pgxpool.Pool, rdb *cache.RedisClient) *CacheHandler {
	return &CacheHandler{DB: db, Cache: rdb}
}

// Invalidate force-purges the cache for a project
func (h *CacheHandler) Invalidate(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	// Get project slug for cache key pattern
	var slug string
	err := h.DB.QueryRow(context.Background(),
		`SELECT slug FROM projects WHERE id = $1`, projectID,
	).Scan(&slug)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	pattern := cache.ProjectCachePattern(slug)
	if err := h.Cache.DeleteByPattern(context.Background(), pattern); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to invalidate cache"})
	}

	return c.JSON(fiber.Map{
		"message": "Cache invalidated",
		"project": slug,
	})
}

// Status returns cache status for a project
func (h *CacheHandler) Status(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	var slug string
	err := h.DB.QueryRow(context.Background(),
		`SELECT slug FROM projects WHERE id = $1`, projectID,
	).Scan(&slug)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	// Check if any cache keys exist for this project
	pattern := cache.ProjectCachePattern(slug)
	iter := h.Cache.Client.Scan(context.Background(), 0, pattern, 100).Iterator()
	cachedKeys := []string{}
	for iter.Next(context.Background()) {
		cachedKeys = append(cachedKeys, iter.Val())
	}

	return c.JSON(fiber.Map{
		"project":     slug,
		"cached":      len(cachedKeys) > 0,
		"cached_keys": len(cachedKeys),
	})
}

// Rebuild force-populates the cache for all languages in a project
func (h *CacheHandler) Rebuild(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership/editor
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	// Get project slug
	var slug string
	err := h.DB.QueryRow(context.Background(),
		`SELECT slug FROM projects WHERE id = $1`, projectID,
	).Scan(&slug)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	// Get all languages
	rows, err := h.DB.Query(context.Background(),
		"SELECT id, code FROM languages WHERE project_id = $1", projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch languages"})
	}
	defer rows.Close()

	type lang struct {
		ID   string
		Code string
	}
	var languages []lang
	for rows.Next() {
		var l lang
		if err := rows.Scan(&l.ID, &l.Code); err == nil {
			languages = append(languages, l)
		}
	}

	// For each language and format, build the cache
	// Note: We're doing this synchronously for simplicity, but for large projects it should be async.
	for _, l := range languages {
		for _, format := range []string{"json", "msgpack"} {
			if err := h.rebuildCacheForLanguage(slug, projectID, l.ID, l.Code, format); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to rebuild cache for " + l.Code})
			}
		}
	}

	return c.JSON(fiber.Map{
		"message":   "Cache rebuilt successfully",
		"project":   slug,
		"languages": len(languages),
	})
}

func (h *CacheHandler) rebuildCacheForLanguage(slug, projectID, langID, langCode, format string) error {
	rows, err := h.DB.Query(context.Background(),
		`SELECT tk.key, t.value
		 FROM translation_keys tk
		 LEFT JOIN translations t ON t.key_id = tk.id AND t.language_id = $2
		 WHERE tk.project_id = $1
		 ORDER BY tk.key`,
		projectID, langID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	flatMap := make(map[string]string)
	for rows.Next() {
		var key string
		var value *string
		if err := rows.Scan(&key, &value); err == nil {
			if value != nil {
				flatMap[key] = *value
			} else {
				flatMap[key] = ""
			}
		}
	}

	nested := buildNestedMap(flatMap)

	var data []byte
	if format == "msgpack" {
		data, err = msgpack.Marshal(nested)
	} else {
		data, err = json.MarshalIndent(nested, "", "  ")
	}

	if err != nil {
		return err
	}

	cacheKey := cache.CacheKey(slug, langCode, format)
	return h.Cache.Set(context.Background(), cacheKey, data, 1*time.Hour)
}
