package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"translate-management/models"
)

// ---- Translations Handler Tests ----

func TestGetTranslations(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "transowner@test.com", "transowner")
	projID := seedProject(t, "Trans Project", "trans-proj", ownerID)
	langID := seedLanguage(t, projID, "en", "English", true)
	keyID := seedKey(t, projID, "welcome", "Welcome msg")
	seedTranslation(t, keyID, langID, "Hello!")

	h := NewTranslationHandler(testDB, nil)
	app := newTestApp("GET", "/projects/:id/translations", h.Get, ownerID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/translations", projID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var entries []models.TranslationEntry
	json.NewDecoder(resp.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "welcome" {
		t.Errorf("expected key 'welcome', got '%s'", entries[0].Key)
	}
	if entries[0].Values[langID] != "Hello!" {
		t.Errorf("expected value 'Hello!', got '%s'", entries[0].Values[langID])
	}
}

func TestGetTranslations_WithEnvFilter(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "transowner2@test.com", "transowner2")
	projID := seedProject(t, "Trans Project", "trans-proj2", ownerID)
	seedLanguage(t, projID, "en", "English", true)
	keyID1 := seedKey(t, projID, "key.in.env", "")
	keyID2 := seedKey(t, projID, "key.not.in.env", "")
	envID := seedEnvironment(t, projID, "production", "")
	assignKeyToEnv(t, keyID1, envID)
	// keyID2 is NOT in production env

	h := NewTranslationHandler(testDB, nil)
	app := newTestApp("GET", "/projects/:id/translations", h.Get, ownerID)

	url := fmt.Sprintf("/projects/%s/translations?env_id=%s", projID, envID)
	resp := doRequest(t, app, "GET", url, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var entries []models.TranslationEntry
	json.NewDecoder(resp.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 env-scoped entry, got %d", len(entries))
	}
	if entries[0].Key != "key.in.env" {
		t.Errorf("expected 'key.in.env', got '%s'", entries[0].Key)
	}
	_ = keyID2 // only in "all" view
}

func TestGetTranslations_NoAccess(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "transowner3@test.com", "transowner3")
	outsiderID := seedUser(t, "transoutsider@test.com", "transoutsider")
	projID := seedProject(t, "Trans Project", "trans-proj3", ownerID)

	h := NewTranslationHandler(testDB, nil)
	app := newTestApp("GET", "/projects/:id/translations", h.Get, outsiderID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/translations", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestBatchUpdateTranslations(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "batchowner@test.com", "batchowner")
	projID := seedProject(t, "Batch Project", "batch-proj", ownerID)
	langID := seedLanguage(t, projID, "en", "English", true)
	keyID := seedKey(t, projID, "greet", "")

	h := NewTranslationHandler(testDB, nil)
	app := newTestApp("PUT", "/projects/:id/translations", h.BatchUpdate, ownerID)

	body := map[string]interface{}{
		"translations": []map[string]string{
			{"key_id": keyID, "language_id": langID, "value": "Hello World"},
		},
	}
	resp := doRequest(t, app, "PUT", fmt.Sprintf("/projects/%s/translations", projID), body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Verify the value was saved
	var val string
	testDB.QueryRow(nil,
		`SELECT value FROM translations WHERE key_id = $1 AND language_id = $2`,
		keyID, langID,
	).Scan(&val)
	if val != "Hello World" {
		t.Errorf("expected 'Hello World' in DB, got '%s'", val)
	}
}

func TestBatchUpdateTranslations_NoPermission(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "batchowner2@test.com", "batchowner2")
	viewerID := seedUser(t, "batchviewer@test.com", "batchviewer")
	projID := seedProject(t, "Batch Project", "batch-proj2", ownerID)
	testDB.Exec(nil, `INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, 'viewer')`, projID, viewerID)
	langID := seedLanguage(t, projID, "en", "English", true)
	keyID := seedKey(t, projID, "greet", "")

	h := NewTranslationHandler(testDB, nil)
	app := newTestApp("PUT", "/projects/:id/translations", h.BatchUpdate, viewerID)

	body := map[string]interface{}{
		"translations": []map[string]string{
			{"key_id": keyID, "language_id": langID, "value": "Hello"},
		},
	}
	resp := doRequest(t, app, "PUT", fmt.Sprintf("/projects/%s/translations", projID), body)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for viewer, got %d", resp.StatusCode)
	}
}
