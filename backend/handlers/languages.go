package handlers

import (
	"context"

	"translate-management/cache"
	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageHandler struct {
	DB    *pgxpool.Pool
	Cache *cache.RedisClient
}

func NewLanguageHandler(db *pgxpool.Pool, rdb *cache.RedisClient) *LanguageHandler {
	return &LanguageHandler{DB: db, Cache: rdb}
}

// List returns all languages for a project
func (h *LanguageHandler) List(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project membership
	var exists bool
	err := h.DB.QueryRow(context.Background(), 
		`SELECT EXISTS(
			SELECT 1 FROM projects p 
			LEFT JOIN project_members pm ON p.id = pm.project_id 
			WHERE p.id = $1 AND (p.created_by = $2 OR pm.user_id = $2)
		)`, projectID, userID).Scan(&exists)
	
	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found or access denied"})
	}

	rows, err := h.DB.Query(context.Background(),
		`SELECT id, project_id, code, name, is_default, created_at 
		 FROM languages WHERE project_id = $1 ORDER BY is_default DESC, name ASC`,
		projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch languages"})
	}
	defer rows.Close()

	languages := []models.Language{}
	for rows.Next() {
		var l models.Language
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Code, &l.Name, &l.IsDefault, &l.CreatedAt); err != nil {
			continue
		}
		languages = append(languages, l)
	}

	return c.JSON(languages)
}

// Create adds a new language to a project
func (h *LanguageHandler) Create(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project membership and role
	var role string
	err := h.DB.QueryRow(context.Background(), 
		`SELECT 
			CASE 
				WHEN p.created_by = $2 THEN 'owner'
				ELSE pm.role
			END as role
		FROM projects p 
		LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
		WHERE p.id = $1`, 
		projectID, userID).Scan(&role)
	
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found or access denied"})
	}

	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var req models.CreateLanguageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Code == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code and name are required"})
	}

	// If setting as default, unset other defaults first
	if req.IsDefault {
		_, _ = h.DB.Exec(context.Background(),
			`UPDATE languages SET is_default = FALSE WHERE project_id = $1`, projectID,
		)
	}

	var l models.Language
	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO languages (project_id, code, name, is_default) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, project_id, code, name, is_default, created_at`,
		projectID, req.Code, req.Name, req.IsDefault,
	).Scan(&l.ID, &l.ProjectID, &l.Code, &l.Name, &l.IsDefault, &l.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create language. Code might already exist."})
	}

	h.invalidateCache(projectID)

	return c.Status(fiber.StatusCreated).JSON(l)
}

// Update updates a language
func (h *LanguageHandler) Update(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project membership and role
	var role string
	err := h.DB.QueryRow(context.Background(), 
		`SELECT 
			CASE 
				WHEN p.created_by = $2 THEN 'owner'
				ELSE pm.role
			END as role
		FROM projects p 
		LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
		WHERE p.id = $1`, 
		projectID, userID).Scan(&role)
	
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found or access denied"})
	}

	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}
	langID := c.Params("langId")

	var req models.UpdateLanguageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.IsDefault {
		_, _ = h.DB.Exec(context.Background(),
			`UPDATE languages SET is_default = FALSE WHERE project_id = $1`, projectID,
		)
	}

	var l models.Language
	err = h.DB.QueryRow(context.Background(),
		`UPDATE languages SET name = $1, is_default = $2 
		 WHERE id = $3 AND project_id = $4 
		 RETURNING id, project_id, code, name, is_default, created_at`,
		req.Name, req.IsDefault, langID, projectID,
	).Scan(&l.ID, &l.ProjectID, &l.Code, &l.Name, &l.IsDefault, &l.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found"})
	}

	h.invalidateCache(projectID)

	return c.JSON(l)
}

// Delete removes a language
func (h *LanguageHandler) Delete(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project membership and role
	var role string
	err := h.DB.QueryRow(context.Background(), 
		`SELECT 
			CASE 
				WHEN p.created_by = $2 THEN 'owner'
				ELSE pm.role
			END as role
		FROM projects p 
		LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
		WHERE p.id = $1`, 
		projectID, userID).Scan(&role)
	
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found or access denied"})
	}

	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}
	langID := c.Params("langId")

	result, err := h.DB.Exec(context.Background(),
		`DELETE FROM languages WHERE id = $1 AND project_id = $2`, langID, projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete language"})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found"})
	}

	h.invalidateCache(projectID)

	return c.JSON(fiber.Map{"message": "Language deleted"})
}

func (h *LanguageHandler) invalidateCache(projectID string) {
	var slug string
	_ = h.DB.QueryRow(context.Background(), "SELECT slug FROM projects WHERE id = $1", projectID).Scan(&slug)
	if slug != "" {
		_ = h.Cache.DeleteByPattern(context.Background(), cache.ProjectCachePattern(slug))
	}
}
