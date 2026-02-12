package handlers

import (
	"context"
	"log"
	"time"
	"translate-management/models"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationHandler struct {
	DB *pgxpool.Pool
}

func NewInvitationHandler(db *pgxpool.Pool) *InvitationHandler {
	return &InvitationHandler{DB: db}
}

// InviteUser sends an invitation to a user
func (h *InvitationHandler) InviteUser(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}
	if req.Role == "" {
		req.Role = "viewer"
	}

	// Verify project ownership
	var projectExists bool
	err := h.DB.QueryRow(context.Background(), 
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND created_by = $2)", 
		projectID, userID).Scan(&projectExists)
	
	if err != nil || !projectExists {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Project not found or you don't have permission to invite"})
	}

	// Check if user is already a member
	var isMember bool
	err = h.DB.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM project_members pm
			JOIN users u ON u.id = pm.user_id
			WHERE pm.project_id = $1 AND u.email = $2
		)`, projectID, req.Email).Scan(&isMember)
	
	if isMember {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User is already a member of this project"})
	}

	// Create invitation
	var invitation models.ProjectInvitation
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days expiration

	err = h.DB.QueryRow(context.Background(),
		`INSERT INTO project_invitations (project_id, email, role, invited_by, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (project_id, email) DO UPDATE SET 
		 	role = EXCLUDED.role, 
			expires_at = EXCLUDED.expires_at,
			status = 'pending'
		 RETURNING id, project_id, email, role, invited_by, status, created_at, expires_at`,
		projectID, req.Email, req.Role, userID, expiresAt,
	).Scan(&invitation.ID, &invitation.ProjectID, &invitation.Email, &invitation.Role, &invitation.InvitedBy, &invitation.Status, &invitation.CreatedAt, &invitation.ExpiresAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create invitation"})
	}

	return c.Status(fiber.StatusCreated).JSON(invitation)
}

// GetInvitations returns pending invitations for the current user (by email)
func (h *InvitationHandler) GetInvitations(c *fiber.Ctx) error {
	userEmail := c.Locals("user_email").(string) // Assuming email is available in Locals from auth middleware
	log.Printf("Fetching invitations for email: %s", userEmail)

	rows, err := h.DB.Query(context.Background(),
		`SELECT i.id, i.project_id, i.email, i.role, i.invited_by, i.status, i.created_at, i.expires_at, p.name as project_name, u.name as inviter_name
		 FROM project_invitations i
		 JOIN projects p ON p.id = i.project_id
		 JOIN users u ON u.id = i.invited_by
		 WHERE i.email = $1 AND i.status = 'pending' AND i.expires_at > NOW()`,
		userEmail,
	)
	if err != nil {
		log.Printf("Error fetching invitations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch invitations"})
	}
	defer rows.Close()

	type InvitationWithDetails struct {
		models.ProjectInvitation
		ProjectName string `json:"project_name"`
		InviterName string `json:"inviter_name"`
	}

	invitations := []InvitationWithDetails{}
	for rows.Next() {
		var i InvitationWithDetails
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.Email, &i.Role, &i.InvitedBy, &i.Status, &i.CreatedAt, &i.ExpiresAt, &i.ProjectName, &i.InviterName); err != nil {
			continue
		}
		invitations = append(invitations, i)
	}

	return c.JSON(invitations)
}

// RespondToInvitation accepts or rejects an invitation
func (h *InvitationHandler) RespondToInvitation(c *fiber.Ctx) error {
	invitationID := c.Params("id")
	userID := c.Locals("user_id").(string)
	userEmail := c.Locals("user_email").(string)

	var req struct {
		Accept bool `json:"accept"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer tx.Rollback(context.Background())

	// Verify invitation
	var inv models.ProjectInvitation
	err = tx.QueryRow(context.Background(),
		`SELECT id, project_id, email, role FROM project_invitations 
		 WHERE id = $1 AND email = $2 AND status = 'pending' AND expires_at > NOW()`,
		invitationID, userEmail).Scan(&inv.ID, &inv.ProjectID, &inv.Email, &inv.Role)
	
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invitation not found or expired"})
	}

	newStatus := "rejected"
	if req.Accept {
		newStatus = "accepted"
		// Add to project members
		_, err := tx.Exec(context.Background(),
			`INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, $3)`,
			inv.ProjectID, userID, inv.Role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add member"})
		}
	}

	// Update invitation status
	_, err = tx.Exec(context.Background(),
		`UPDATE project_invitations SET status = $1 WHERE id = $2`,
		newStatus, invitationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update invitation"})
	}

	if err := tx.Commit(context.Background()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction failed"})
	}

	return c.JSON(fiber.Map{"message": "Invitation " + newStatus})
}
