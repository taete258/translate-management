package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"translate-management/models"

	"github.com/gofiber/fiber/v2"
)

// newTestApp creates a minimal Fiber app with the provided handler and locals injected.
func newTestApp(method, path string, handler fiber.Handler, userID string) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	}
	return app
}

// doRequest sends a test request and returns the response.
func doRequest(t *testing.T, app *fiber.App, method, url string, body interface{}) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, url, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return resp
}

// ---- Environments Handler Tests ----

func TestListEnvironments(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner@test.com", "owner")
	projID := seedProject(t, "Test Project", "test-proj", ownerID)
	seedEnvironment(t, projID, "production", "Prod")
	seedEnvironment(t, projID, "staging", "Staging")

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("GET", "/projects/:id/environments", h.List, ownerID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/environments", projID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var envs []models.Environment
	json.NewDecoder(resp.Body).Decode(&envs)
	if len(envs) != 2 {
		t.Errorf("expected 2 environments, got %d", len(envs))
	}
}

func TestListEnvironments_NotMember(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner2@test.com", "owner2")
	otherID := seedUser(t, "other@test.com", "other")
	projID := seedProject(t, "Test Project", "test-proj2", ownerID)

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("GET", "/projects/:id/environments", h.List, otherID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/environments", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateEnvironment(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner3@test.com", "owner3")
	projID := seedProject(t, "Test Project", "test-proj3", ownerID)

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("POST", "/projects/:id/environments", h.Create, ownerID)

	body := map[string]string{"name": "production", "description": "Prod env"}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/environments", projID), body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var env models.Environment
	json.NewDecoder(resp.Body).Decode(&env)
	if env.Name != "production" {
		t.Errorf("expected name 'production', got '%s'", env.Name)
	}
	if env.ProjectID != projID {
		t.Errorf("expected project_id '%s', got '%s'", projID, env.ProjectID)
	}
}

func TestCreateEnvironment_DuplicateName(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner4@test.com", "owner4")
	projID := seedProject(t, "Test Project", "test-proj4", ownerID)
	seedEnvironment(t, projID, "production", "existing")

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("POST", "/projects/:id/environments", h.Create, ownerID)

	body := map[string]string{"name": "production"}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/environments", projID), body)
	if resp.StatusCode == http.StatusCreated {
		t.Fatal("expected error on duplicate env name, got 201")
	}
}

func TestCreateEnvironment_NotOwnerOrEditor(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner5@test.com", "owner5")
	viewerID := seedUser(t, "viewer@test.com", "viewer5")
	projID := seedProject(t, "Test Project", "test-proj5", ownerID)
	// Add viewer as project member with 'viewer' role
	testDB.Exec(nil, `INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, 'viewer')`, projID, viewerID)

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("POST", "/projects/:id/environments", h.Create, viewerID)

	body := map[string]string{"name": "staging"}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/environments", projID), body)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer, got %d", resp.StatusCode)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner6@test.com", "owner6")
	projID := seedProject(t, "Test Project", "test-proj6", ownerID)
	envID := seedEnvironment(t, projID, "staging", "")

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("DELETE", "/projects/:id/environments/:envId", h.Delete, ownerID)

	resp := doRequest(t, app, "DELETE", fmt.Sprintf("/projects/%s/environments/%s", projID, envID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteEnvironment_NotFound(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner7@test.com", "owner7")
	projID := seedProject(t, "Test Project", "test-proj7", ownerID)

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("DELETE", "/projects/:id/environments/:envId", h.Delete, ownerID)

	resp := doRequest(t, app, "DELETE",
		fmt.Sprintf("/projects/%s/environments/00000000-0000-0000-0000-000000000000", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateEnvironment_CloneKeys(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "owner8@test.com", "owner8")
	projID := seedProject(t, "Test Project", "test-proj8", ownerID)
	// Seed 2 keys
	seedKey(t, projID, "key.one", "Key One")
	seedKey(t, projID, "key.two", "Key Two")

	h := NewEnvironmentHandler(testDB)
	app := newTestApp("POST", "/projects/:id/environments", h.Create, ownerID)

	// Create env with clone_keys = true
	body := map[string]interface{}{
		"name":        "staging",
		"description": "Staging Env",
		"clone_keys":  true,
	}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/environments", projID), body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var env models.Environment
	json.NewDecoder(resp.Body).Decode(&env)

	// Verify keys were cloned
	var count int
	err := testDB.QueryRow(nil, `SELECT COUNT(*) FROM key_environments WHERE env_id = $1`, env.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query key_environments: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 keys cloned, got %d", count)
	}
}
