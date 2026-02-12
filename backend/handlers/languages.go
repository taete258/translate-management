package handlers

import (
	"context"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageHandler struct {
	DB *pgxpool.Pool
}

func NewLanguageHandler(db *pgxpool.Pool) *LanguageHandler {
	return &LanguageHandler{DB: db}
}

// List returns all languages for a project
func (h *LanguageHandler) List(c *fiber.Ctx) error {
	projectID := c.Params("id")

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
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO languages (project_id, code, name, is_default) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, project_id, code, name, is_default, created_at`,
		projectID, req.Code, req.Name, req.IsDefault,
	).Scan(&l.ID, &l.ProjectID, &l.Code, &l.Name, &l.IsDefault, &l.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create language. Code might already exist."})
	}

	return c.Status(fiber.StatusCreated).JSON(l)
}

// Update updates a language
func (h *LanguageHandler) Update(c *fiber.Ctx) error {
	projectID := c.Params("id")
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
	err := h.DB.QueryRow(context.Background(),
		`UPDATE languages SET name = $1, is_default = $2 
		 WHERE id = $3 AND project_id = $4 
		 RETURNING id, project_id, code, name, is_default, created_at`,
		req.Name, req.IsDefault, langID, projectID,
	).Scan(&l.ID, &l.ProjectID, &l.Code, &l.Name, &l.IsDefault, &l.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found"})
	}

	return c.JSON(l)
}

// Delete removes a language
func (h *LanguageHandler) Delete(c *fiber.Ctx) error {
	projectID := c.Params("id")
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

	return c.JSON(fiber.Map{"message": "Language deleted"})
}
