package models

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after successful login/register
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateProjectRequest is the request body for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

// UpdateProjectRequest is the request body for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

// CreateLanguageRequest is the request body for adding a language
type CreateLanguageRequest struct {
	Code      string `json:"code" validate:"required,min=2,max=10"`
	Name      string `json:"name" validate:"required,min=1,max=100"`
	IsDefault bool   `json:"is_default"`
}

// UpdateLanguageRequest is the request body for updating a language
type UpdateLanguageRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	IsDefault bool   `json:"is_default"`
}

// CreateKeyRequest is the request body for creating a translation key
type CreateKeyRequest struct {
	Key         string `json:"key" validate:"required,min=1,max=500"`
	Description string `json:"description"`
}

// UpdateKeyRequest is the request body for updating a translation key
type UpdateKeyRequest struct {
	Key         string `json:"key" validate:"required,min=1,max=500"`
	Description string `json:"description"`
}

// BatchTranslationUpdate represents a batch of translation updates
type BatchTranslationUpdate struct {
	Translations []TranslationUpdate `json:"translations" validate:"required"`
}

// TranslationUpdate represents a single translation value update
type TranslationUpdate struct {
	KeyID      string `json:"key_id" validate:"required"`
	LanguageID string `json:"language_id" validate:"required"`
	Value      string `json:"value"`
}

// CreateAPIKeyRequest is the request body for generating an API key
type CreateAPIKeyRequest struct {
	Name   string   `json:"name" validate:"required,min=1,max=255"`
	Scopes []string `json:"scopes"`
}

// CreateAPIKeyResponse includes the raw key (only shown once)
type CreateAPIKeyResponse struct {
	APIKey APIKey `json:"api_key"`
	RawKey string `json:"raw_key"`
}

// ImportRequest for importing translation JSON
type ImportRequest struct {
	LanguageCode string            `json:"language_code" validate:"required"`
	Translations map[string]interface{} `json:"translations" validate:"required"`
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginationQuery holds pagination parameters
type PaginationQuery struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Search string `query:"search"`
}
