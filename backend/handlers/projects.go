package handlers

import (
	"context"
	"regexp"
	"strings"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectHandler struct {
	DB *pgxpool.Pool
}

func NewProjectHandler(db *pgxpool.Pool) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

// List returns all projects
func (h *ProjectHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	search := c.Query("search", "")

	var rows interface{ Close() }
	query := `SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects WHERE created_by = $1`
	args := []interface{}{userID}

	if search != "" {
		query += ` AND (name ILIKE $2 OR slug ILIKE $2)`
		args = append(args, "%"+search+"%")
	}
	query += ` ORDER BY created_at DESC`

	r, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch projects"})
	}
	rows = r
	defer rows.(interface{ Close() }).Close()

	projects := []models.Project{}
	for r.Next() {
		var p models.Project
		if err := r.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		projects = append(projects, p)
	}

	return c.JSON(projects)
}

// Get returns a single project by ID
func (h *ProjectHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var p models.Project
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects WHERE id = $1 AND created_by = $2`,
		id, userID,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	return c.JSON(p)
}

// Create creates a new project
func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req models.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	slug := generateSlug(req.Name)

	var p models.Project
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO projects (name, slug, description, created_by) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, name, slug, description, created_by, created_at, updated_at`,
		req.Name, slug, req.Description, userID,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create project"})
	}

	return c.Status(fiber.StatusCreated).JSON(p)
}

// Update updates a project
func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req models.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	var p models.Project
	err := h.DB.QueryRow(context.Background(),
		`UPDATE projects SET name = $1, description = $2, updated_at = NOW() 
		 WHERE id = $3 AND created_by = $4
		 RETURNING id, name, slug, description, created_by, created_at, updated_at`,
		req.Name, req.Description, id, userID,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	return c.JSON(p)
}

// Delete removes a project
func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	result, err := h.DB.Exec(context.Background(), `DELETE FROM projects WHERE id = $1 AND created_by = $2`, id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete project"})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	return c.JSON(fiber.Map{"message": "Project deleted"})
}

// Stats returns project statistics
func (h *ProjectHandler) Stats(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Validate ownership
	var exists bool
	err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", id, userID).Scan(&exists)
	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}

	var totalKeys, totalLangs int
	_ = h.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM translation_keys WHERE project_id = $1`, id,
	).Scan(&totalKeys)
	_ = h.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM languages WHERE project_id = $1`, id,
	).Scan(&totalLangs)

	// Language progress
	progress := make(map[string]float64)
	if totalKeys > 0 {
		rows, err := h.DB.Query(context.Background(),
			`SELECT l.code, COUNT(t.id) 
			 FROM languages l 
			 LEFT JOIN translations t ON t.language_id = l.id AND t.value != ''
			 WHERE l.project_id = $1 
			 GROUP BY l.code`, id)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var code string
				var count int
				if rows.Scan(&code, &count) == nil {
					progress[code] = float64(count) / float64(totalKeys) * 100
				}
			}
		}
	}

	return c.JSON(models.ProjectStats{
		TotalKeys:        totalKeys,
		TotalLanguages:   totalLangs,
		LanguageProgress: progress,
	})
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
