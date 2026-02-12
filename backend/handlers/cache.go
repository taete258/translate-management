package handlers

import (
	"context"

	"translate-management/cache"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
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
