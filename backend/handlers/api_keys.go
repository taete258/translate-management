package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKeyHandler struct {
	DB *pgxpool.Pool
}

func NewAPIKeyHandler(db *pgxpool.Pool) *APIKeyHandler {
	return &APIKeyHandler{DB: db}
}

// List returns all API keys for a project
func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	projectID := c.Params("id")

	rows, err := h.DB.Query(context.Background(),
		`SELECT id, project_id, name, key_prefix, scopes, is_active, last_used_at, created_at 
		 FROM api_keys WHERE project_id = $1 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch API keys"})
	}
	defer rows.Close()

	keys := []models.APIKey{}
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Name, &k.KeyPrefix, &k.Scopes, &k.IsActive, &k.LastUsedAt, &k.CreatedAt); err != nil {
			continue
		}
		keys = append(keys, k)
	}

	return c.JSON(keys)
}

// Create generates a new API key
func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	projectID := c.Params("id")

	var req models.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}

	if len(req.Scopes) == 0 {
		req.Scopes = []string{"read"}
	}

	// Generate random API key
	rawKey, err := generateAPIKey()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate key"})
	}

	// Hash for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := fmt.Sprintf("%x", hash)
	keyPrefix := rawKey[:8]

	var k models.APIKey
	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO api_keys (project_id, name, key_hash, key_prefix, scopes) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id, project_id, name, key_prefix, scopes, is_active, last_used_at, created_at`,
		projectID, req.Name, keyHash, keyPrefix, req.Scopes,
	).Scan(&k.ID, &k.ProjectID, &k.Name, &k.KeyPrefix, &k.Scopes, &k.IsActive, &k.LastUsedAt, &k.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create API key"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.CreateAPIKeyResponse{
		APIKey: k,
		RawKey: rawKey,
	})
}

// Delete deactivates an API key
func (h *APIKeyHandler) Delete(c *fiber.Ctx) error {
	projectID := c.Params("id")
	keyID := c.Params("keyId")

	result, err := h.DB.Exec(context.Background(),
		`UPDATE api_keys SET is_active = FALSE WHERE id = $1 AND project_id = $2`,
		keyID, projectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to deactivate API key"})
	}

	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "API key not found"})
	}

	return c.JSON(fiber.Map{"message": "API key deactivated"})
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "tm_" + hex.EncodeToString(bytes), nil
}
