package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImportHandler struct {
	DB *pgxpool.Pool
}

func NewImportHandler(db *pgxpool.Pool) *ImportHandler {
	return &ImportHandler{DB: db}
}

// Import imports translation JSON data into a project
func (h *ImportHandler) Import(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req models.ImportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.LanguageCode == "" || req.Translations == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Language code and translations are required"})
	}

	// Get or create language
	var langID string
	err := h.DB.QueryRow(context.Background(),
		`SELECT id FROM languages WHERE project_id = $1 AND code = $2`,
		projectID, req.LanguageCode,
	).Scan(&langID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Language not found. Create it first."})
	}

	// Flatten nested JSON
	flat := make(map[string]string)
	flattenJSON("", req.Translations, flat)

	tx, err := h.DB.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to begin transaction"})
	}
	defer tx.Rollback(context.Background())

	imported := 0
	for key, value := range flat {
		// Upsert key
		var keyID string
		err := tx.QueryRow(context.Background(),
			`INSERT INTO translation_keys (project_id, key) 
			 VALUES ($1, $2) 
			 ON CONFLICT (project_id, key) DO UPDATE SET updated_at = NOW()
			 RETURNING id`,
			projectID, key,
		).Scan(&keyID)

		if err != nil {
			continue
		}

		// Upsert translation
		_, err = tx.Exec(context.Background(),
			`INSERT INTO translations (key_id, language_id, value, updated_by) 
			 VALUES ($1, $2, $3, $4) 
			 ON CONFLICT (key_id, language_id) 
			 DO UPDATE SET value = EXCLUDED.value, updated_at = NOW(), updated_by = EXCLUDED.updated_by`,
			keyID, langID, value, userID,
		)

		if err == nil {
			imported++
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit import"})
	}

	return c.JSON(fiber.Map{
		"message":  "Import completed",
		"imported": imported,
	})
}

// flattenJSON converts nested maps to dot-notation flat keys
func flattenJSON(prefix string, data map[string]interface{}, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			flattenJSON(fullKey, v, result)
		case string:
			result[fullKey] = v
		case json.Number:
			result[fullKey] = v.String()
		case float64:
			result[fullKey] = strings.TrimRight(strings.TrimRight(
				strings.Replace(
					strings.Replace(
						formatFloat(v), "e+", "e", 1),
					"e-", "e-", 1),
				"0"), ".")
		default:
			// Convert to string
			if b, err := json.Marshal(v); err == nil {
				result[fullKey] = string(b)
			}
		}
	}
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(
			strings.Replace(
				json.Number(strings.TrimRight(strings.TrimRight(
					func() string { b, _ := json.Marshal(f); return string(b) }(),
					"0"), ".")).String(),
				"e+", "e", 1),
			"e-", "e-", 1),
		"0"), ".")
}
