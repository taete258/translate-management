package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vmihailenco/msgpack/v5"
)

type ProjectExportHandler struct {
	DB *pgxpool.Pool
}

func NewProjectExportHandler(db *pgxpool.Pool) *ProjectExportHandler {
	return &ProjectExportHandler{DB: db}
}

// ExportLanguage exports translations for a specific language in a project (JWT protected)
func (h *ProjectExportHandler) ExportLanguage(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Verify project ownership
	var exists bool
	if err := h.DB.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", projectID, userID).Scan(&exists); err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Project not found"})
	}
	langCode := c.Params("langCode")
	format := c.Query("format", "json")

	if format != "json" && format != "msgpack" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format must be 'json' or 'msgpack'"})
	}

	// Get language ID
	var languageID string
	err := h.DB.QueryRow(context.Background(),
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
		c.Set("Content-Disposition", "attachment; filename="+langCode+".msgpack")
	} else {
		data, err = json.MarshalIndent(nested, "", "  ")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to encode JSON"})
		}
		c.Set("Content-Type", "application/json")
		c.Set("Content-Disposition", "attachment; filename="+langCode+".json")
	}

	return c.Send(data)
}
