package models

import "time"

// User represents a user account
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Project represents a translation project
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectWithRole includes the user's role in the project
type ProjectWithRole struct {
	Project
	Role string `json:"role"`
}

// Language represents a language within a project
type Language struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// TranslationKey represents a translation key within a project
type TranslationKey struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Translation represents a translated value
type Translation struct {
	ID         string    `json:"id"`
	KeyID      string    `json:"key_id"`
	LanguageID string    `json:"language_id"`
	Value      string    `json:"value"`
	UpdatedAt  time.Time `json:"updated_at"`
	UpdatedBy  *string   `json:"updated_by,omitempty"`
}

// APIKey represents an API key for external access
type APIKey struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	Name       string    `json:"name"`
	KeyHash    string    `json:"-"`
	KeyPrefix  string    `json:"key_prefix"`
	Scopes     []string  `json:"scopes"`
	IsActive   bool      `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// TranslationEntry is used for the translation grid (key + all language values)
type TranslationEntry struct {
	KeyID       string            `json:"key_id"`
	Key         string            `json:"key"`
	Description string            `json:"description"`
	Values      map[string]string `json:"values"` // language_id -> value
}

// ProjectStats holds project statistics
type ProjectStats struct {
	TotalKeys       int                `json:"total_keys"`
	TotalLanguages  int                `json:"total_languages"`
	LanguageProgress map[string]float64 `json:"language_progress"` // language_code -> percentage
}

// ProjectMember represents a user who is a member of a project
type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ProjectInvitation represents an invitation to join a project
type ProjectInvitation struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	InvitedBy string    `json:"invited_by"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ProjectMemberInfo includes user details for a project member
type ProjectMemberInfo struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
}
