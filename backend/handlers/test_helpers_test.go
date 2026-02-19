// Package handlers provides HTTP handlers for the translate management API.
// This file contains shared test setup and helper utilities for handler tests.
//
// Tests require a PostgreSQL database. Set TEST_DATABASE_URL to run them:
//
//	TEST_DATABASE_URL="postgres://user:pass@localhost:5432/testdb" go test ./handlers/...
//
// If TEST_DATABASE_URL is not set, integration tests are skipped automatically.
package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testDB holds the shared DB pool for handler tests.
var testDB *pgxpool.Pool

// TestMain sets up a real DB connection for tests when TEST_DATABASE_URL is set.
// All tables are seeded fresh for each test via helper functions.
func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn != "" {
		pool, err := pgxpool.New(context.Background(), dsn)
		if err == nil {
			testDB = pool
			applySchema(pool)
		}
	}
	code := m.Run()
	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

// applySchema runs the migrations on the test DB to ensure schema is current.
func applySchema(db *pgxpool.Pool) {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL DEFAULT '',
			avatar_url TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) UNIQUE NOT NULL,
			description TEXT DEFAULT '',
			created_by UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_members (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'viewer',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS languages (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			code VARCHAR(10) NOT NULL,
			name VARCHAR(100) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, code)
		)`,
		`CREATE TABLE IF NOT EXISTS translation_keys (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			key VARCHAR(500) NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, key)
		)`,
		`CREATE TABLE IF NOT EXISTS translations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			key_id UUID NOT NULL REFERENCES translation_keys(id) ON DELETE CASCADE,
			language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
			value TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
			UNIQUE(key_id, language_id)
		)`,
		`CREATE TABLE IF NOT EXISTS environments (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			description TEXT DEFAULT '',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS key_environments (
			key_id UUID NOT NULL REFERENCES translation_keys(id) ON DELETE CASCADE,
			env_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
			PRIMARY KEY (key_id, env_id)
		)`,
	}
	for _, sql := range migrations {
		db.Exec(context.Background(), sql)
	}
}

// cleanDB truncates all tables between tests for test isolation.
func cleanDB(t *testing.T) {
	t.Helper()
	if testDB == nil {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	tables := []string{
		"key_environments", "environments", "translations",
		"translation_keys", "languages", "project_members", "projects", "users",
	}
	for _, tbl := range tables {
		testDB.Exec(context.Background(), "DELETE FROM "+tbl)
	}
}

// seedUser inserts a test user and returns their ID.
func seedUser(t *testing.T, email, username string) string {
	t.Helper()
	var id string
	err := testDB.QueryRow(context.Background(),
		`INSERT INTO users (email, username, password_hash, name)
		 VALUES ($1, $2, 'hash', 'Test User')
		 ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email
		 RETURNING id`,
		email, username,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	return id
}

// seedProject inserts a test project owned by ownerID and returns the project ID.
func seedProject(t *testing.T, name, slug, ownerID string) string {
	t.Helper()
	var id string
	err := testDB.QueryRow(context.Background(),
		`INSERT INTO projects (name, slug, description, created_by)
		 VALUES ($1, $2, '', $3)
		 RETURNING id`,
		name, slug, ownerID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedProject: %v", err)
	}
	return id
}

// seedLanguage inserts a test language and returns its ID.
func seedLanguage(t *testing.T, projectID, code, name string, isDefault bool) string {
	t.Helper()
	var id string
	err := testDB.QueryRow(context.Background(),
		`INSERT INTO languages (project_id, code, name, is_default)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		projectID, code, name, isDefault,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedLanguage: %v", err)
	}
	return id
}

// seedKey inserts a translation key and returns its ID.
func seedKey(t *testing.T, projectID, key, desc string) string {
	t.Helper()
	var id string
	err := testDB.QueryRow(context.Background(),
		`INSERT INTO translation_keys (project_id, key, description)
		 VALUES ($1, $2, $3) RETURNING id`,
		projectID, key, desc,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedKey: %v", err)
	}
	return id
}

// seedTranslation inserts a translation value.
func seedTranslation(t *testing.T, keyID, langID, value string) {
	t.Helper()
	_, err := testDB.Exec(context.Background(),
		`INSERT INTO translations (key_id, language_id, value)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key_id, language_id) DO UPDATE SET value = EXCLUDED.value`,
		keyID, langID, value,
	)
	if err != nil {
		t.Fatalf("seedTranslation: %v", err)
	}
}

// seedEnvironment inserts a project environment and returns its ID.
func seedEnvironment(t *testing.T, projectID, name, desc string) string {
	t.Helper()
	var id string
	err := testDB.QueryRow(context.Background(),
		`INSERT INTO environments (project_id, name, description)
		 VALUES ($1, $2, $3) RETURNING id`,
		projectID, name, desc,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seedEnvironment: %v", err)
	}
	return id
}

// assignKeyToEnv links a key to an environment.
func assignKeyToEnv(t *testing.T, keyID, envID string) {
	t.Helper()
	_, err := testDB.Exec(context.Background(),
		`INSERT INTO key_environments (key_id, env_id) VALUES ($1, $2)`,
		keyID, envID,
	)
	if err != nil {
		t.Fatalf("assignKeyToEnv: %v", err)
	}
}
