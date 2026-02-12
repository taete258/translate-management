package handlers

import (
	"context"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KeyHandler struct {
	DB *pgxpool.Pool
}

func NewKeyHandler(db *pgxpool.Pool) *KeyHandler {
	return &KeyHandler{DB: db}
}

// List returns all translation keys for a project
func (h *KeyHandler) List(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}
	search := c.Query("search", "")

	query := `SELECT id, project_id, key, description, created_at, updated_at 
			  FROM translation_keys WHERE project_id = $1`
	args := []interface{}{projectID}

	if search != "" {
		query += ` AND key ILIKE $2`
		args = append(args, "%"+search+"%")
	}
	query += ` ORDER BY key ASC`

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch keys"})
	}
	defer rows.Close()

	keys := []models.TranslationKey{}
	for rows.Next() {
		var k models.TranslationKey
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Key, &k.Description, &k.CreatedAt, &k.UpdatedAt); err != nil {
			continue
		}
		keys = append(keys, k)
	}

	return c.JSON(keys)
}

// Create adds a new translation key
func (h *KeyHandler) Create(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	var req models.CreateKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Key is required"})
	}

	var k models.TranslationKey
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO translation_keys (project_id, key, description) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, project_id, key, description, created_at, updated_at`,
		projectID, req.Key, req.Description,
	).Scan(&k.ID, &k.ProjectID, &k.Key, &k.Description, &k.CreatedAt, &k.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create key. Key might already exist."})
	}

	return c.Status(fiber.StatusCreated).JSON(k)
}

// Update updates a translation key
func (h *KeyHandler) Update(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}
	keyID := c.Params("keyId")

	var req models.UpdateKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Key is required"})
	}

	var k models.TranslationKey
	err := h.DB.QueryRow(context.Background(),
		`UPDATE translation_keys SET key = $1, description = $2, updated_at = NOW() 
		 WHERE id = $3 AND project_id = $4 
		 RETURNING id, project_id, key, description, created_at, updated_at`,
		req.Key, req.Description, keyID, projectID,
	).Scan(&k.ID, &k.ProjectID, &k.Key, &k.Description, &k.CreatedAt, &k.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Key not found"})
	}

	return c.JSON(k)
}

// Delete removes a translation key
func (h *KeyHandler) Delete(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}
	keyID := c.Params("keyId")

	result, err := h.DB.Exec(context.Background(),
		`DELETE FROM translation_keys WHERE id = $1 AND project_id = $2`, keyID, projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete key"})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Key not found"})
	}

	return c.JSON(fiber.Map{"message": "Key deleted"})
}
