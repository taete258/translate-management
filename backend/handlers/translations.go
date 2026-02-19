package handlers

import (
	"context"
	"fmt"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TranslationHandler struct {
	DB *pgxpool.Pool
}

func NewTranslationHandler(db *pgxpool.Pool) *TranslationHandler {
	return &TranslationHandler{DB: db}
}

// Get returns all translations for a project as a grid
func (h *TranslationHandler) Get(c *fiber.Ctx) error {
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
	search := c.Query("search", "")
	envID := c.Query("env_id", "")

	// Get all keys (optionally filtered by env)
	keyQuery := `SELECT id, key, description FROM translation_keys WHERE project_id = $1`
	keyArgs := []interface{}{projectID}
	argIdx := 2
	if envID != "" {
		keyQuery += ` AND id IN (SELECT key_id FROM key_environments WHERE env_id = $` + fmt.Sprint(argIdx) + `)`
		keyArgs = append(keyArgs, envID)
		argIdx++
	}
	if search != "" {
		keyQuery += ` AND key ILIKE $` + fmt.Sprint(argIdx)
		keyArgs = append(keyArgs, "%"+search+"%")
	}
	keyQuery += ` ORDER BY key ASC`

	keyRows, err := h.DB.Query(context.Background(), keyQuery, keyArgs...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch keys"})
	}
	defer keyRows.Close()

	entries := []models.TranslationEntry{}
	keyIDs := []string{}

	for keyRows.Next() {
		var keyID, key, desc string
		if err := keyRows.Scan(&keyID, &key, &desc); err != nil {
			continue
		}
		entries = append(entries, models.TranslationEntry{
			KeyID:       keyID,
			Key:         key,
			Description: desc,
			Values:      make(map[string]string),
		})
		keyIDs = append(keyIDs, keyID)
	}

	if len(keyIDs) == 0 {
		return c.JSON(entries)
	}

	// Get all translations for these keys
	tRows, err := h.DB.Query(context.Background(),
		`SELECT t.key_id, t.language_id, t.value
		 FROM translations t
		 JOIN translation_keys tk ON t.key_id = tk.id
		 WHERE tk.project_id = $1`,
		projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch translations"})
	}
	defer tRows.Close()

	// Build a map for quick lookup
	translationMap := make(map[string]map[string]string) // key_id -> language_id -> value
	for tRows.Next() {
		var keyID, langID, value string
		if err := tRows.Scan(&keyID, &langID, &value); err != nil {
			continue
		}
		if translationMap[keyID] == nil {
			translationMap[keyID] = make(map[string]string)
		}
		translationMap[keyID][langID] = value
	}

	// Merge translations into entries
	for i := range entries {
		if vals, ok := translationMap[entries[i].KeyID]; ok {
			entries[i].Values = vals
		}
	}

	return c.JSON(entries)
}

// BatchUpdate updates multiple translations at once
func (h *TranslationHandler) BatchUpdate(c *fiber.Ctx) error {
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

	var req models.BatchTranslationUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Translations) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No translations provided"})
	}

	tx, err := h.DB.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to begin transaction"})
	}
	defer tx.Rollback(context.Background())

	for _, t := range req.Translations {
		_, err := tx.Exec(context.Background(),
			`INSERT INTO translations (key_id, language_id, value, updated_by) 
			 VALUES ($1, $2, $3, $4) 
			 ON CONFLICT (key_id, language_id) 
			 DO UPDATE SET value = EXCLUDED.value, updated_at = NOW(), updated_by = EXCLUDED.updated_by`,
			t.KeyID, t.LanguageID, t.Value, userID,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update translation"})
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Translations updated", "count": len(req.Translations)})
}
