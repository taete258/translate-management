package handlers

import (
	"context"

	"translate-management/config"
	"translate-management/middleware"
	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func NewAuthHandler(db *pgxpool.Pool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

// Register creates a new user account
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email, username, and password are required"})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	name := req.Name
	if name == "" {
		name = req.Username
	}

	var user models.User
	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO users (email, username, password_hash, name) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, email, username, name, avatar_url, created_at, updated_at`,
		req.Email, req.Username, string(hash), name,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Name, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` ||
			err.Error() == `ERROR: duplicate key value violates unique constraint "users_username_key" (SQLSTATE 23505)` {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email or username already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Email, h.Cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login authenticates a user
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username and password are required"})
	}

	var user models.User
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, email, username, password_hash, name, avatar_url, created_at, updated_at 
		 FROM users WHERE username = $1`,
		req.Username,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Email, h.Cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var user models.User
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, email, username, name, avatar_url, created_at, updated_at 
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.Name, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// Logout is a placeholder (JWT is stateless, client discards token)
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}
