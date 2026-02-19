package handlers

import (
	"context"
	"fmt"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnvironmentHandler struct {
	DB *pgxpool.Pool
}

func NewEnvironmentHandler(db *pgxpool.Pool) *EnvironmentHandler {
	return &EnvironmentHandler{DB: db}
}

// verifyProjectRole checks project membership and returns the user's role.
// Returns "" and writes a 404 if not found.
func (h *EnvironmentHandler) verifyProjectRole(c *fiber.Ctx, projectID, userID string) (string, error) {
	var role string
	err := h.DB.QueryRow(context.Background(),
		`SELECT
			CASE
				WHEN p.created_by = $2 THEN 'owner'
				ELSE COALESCE(pm.role, '')
			END as role
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
		WHERE p.id = $1`,
		projectID, userID).Scan(&role)
	if err != nil {
		c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found or access denied"})
		return "", err
	}
	return role, nil
}

// List returns all environments for a project
func (h *EnvironmentHandler) List(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify membership (any role)
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
		`SELECT id, project_id, name, description, created_at
		 FROM environments WHERE project_id = $1 ORDER BY name ASC`,
		projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch environments"})
	}
	defer rows.Close()

	envs := []models.Environment{}
	for rows.Next() {
		var e models.Environment
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.Description, &e.CreatedAt); err != nil {
			continue
		}
		envs = append(envs, e)
	}

	return c.JSON(envs)
}

// Create adds a new environment to a project
func (h *EnvironmentHandler) Create(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	role, err := h.verifyProjectRole(c, projectID, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var req models.CreateEnvironmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Environment name is required"})
	}

	var env models.Environment
	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO environments (project_id, name, description)
		 VALUES ($1, $2, $3)
		 RETURNING id, project_id, name, description, created_at`,
		projectID, req.Name, req.Description,
	).Scan(&env.ID, &env.ProjectID, &env.Name, &env.Description, &env.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create environment. Name might already exist."})
	}

	// Clone keys if requested
	if req.CloneKeys {
		_, err := h.DB.Exec(context.Background(),
			`INSERT INTO key_environments (key_id, env_id)
			 SELECT id, $1 FROM translation_keys WHERE project_id = $2`,
			env.ID, projectID,
		)
		if err != nil {
			// Log error but don't fail the request completely
			// Ideally we should have better logging here
			fmt.Printf("Failed to clone keys for env %s: %v\n", env.ID, err)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(env)
}

// Update modifies an existing environment
func (h *EnvironmentHandler) Update(c *fiber.Ctx) error {
	projectID := c.Params("id")
	envID := c.Params("envId")
	userID := c.Locals("user_id").(string)

	role, err := h.verifyProjectRole(c, projectID, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var req models.UpdateEnvironmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Environment name is required"})
	}

	commandTag, err := h.DB.Exec(context.Background(),
		`UPDATE environments SET name = $1, description = $2 WHERE id = $3 AND project_id = $4`,
		req.Name, req.Description, envID, projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update environment"})
	}
	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Environment not found"})
	}

	var env models.Environment
	err = h.DB.QueryRow(context.Background(),
		`SELECT id, project_id, name, description, created_at FROM environments WHERE id = $1`,
		envID,
	).Scan(&env.ID, &env.ProjectID, &env.Name, &env.Description, &env.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated environment"})
	}

	return c.JSON(env)
}

// Delete removes an environment from a project
func (h *EnvironmentHandler) Delete(c *fiber.Ctx) error {
	projectID := c.Params("id")
	envID := c.Params("envId")
	userID := c.Locals("user_id").(string)

	role, err := h.verifyProjectRole(c, projectID, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	result, err := h.DB.Exec(context.Background(),
		`DELETE FROM environments WHERE id = $1 AND project_id = $2`,
		envID, projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete environment"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Environment not found"})
	}

	return c.JSON(fiber.Map{"message": "Environment deleted"})
}
