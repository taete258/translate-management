package handlers

import (
	"fmt"
	"net/http"
	"testing"
)

// ---- ProjectExport Handler Tests ----

func TestExportLanguage_NoEnv(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "exportowner@test.com", "exportowner")
	projID := seedProject(t, "Export Project", "export-proj", ownerID)
	langID := seedLanguage(t, projID, "en", "English", true)
	keyID := seedKey(t, projID, "app.title", "App title")
	seedTranslation(t, keyID, langID, "My App")

	h := NewProjectExportHandler(testDB)
	app := newTestApp("GET", "/projects/:id/export/:langCode", h.ExportLanguage, ownerID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/export/en", projID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Response should be JSON and contain "My App"
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestExportLanguage_WithEnv(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "exportowner2@test.com", "exportowner2")
	projID := seedProject(t, "Export Project", "export-proj2", ownerID)
	langID := seedLanguage(t, projID, "en", "English", true)
	keyID1 := seedKey(t, projID, "prod.key", "In prod")
	keyID2 := seedKey(t, projID, "other.key", "Not in prod")
	seedTranslation(t, keyID1, langID, "Prod Value")
	seedTranslation(t, keyID2, langID, "Other Value")
	envID := seedEnvironment(t, projID, "production", "")
	assignKeyToEnv(t, keyID1, envID)
	// keyID2 NOT in production env

	h := NewProjectExportHandler(testDB)
	app := newTestApp("GET", "/projects/:id/export/:langCode", h.ExportLanguage, ownerID)

	url := fmt.Sprintf("/projects/%s/export/en?env_id=%s", projID, envID)
	resp := doRequest(t, app, "GET", url, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// The exported JSON should NOT contain "Other Value"
	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])

	if !contains(body, "Prod Value") {
		t.Errorf("expected 'Prod Value' in export body")
	}
	if contains(body, "Other Value") {
		t.Errorf("expected 'Other Value' to be excluded from env-filtered export")
	}
}

func TestExportLanguage_LangNotFound(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "exportowner3@test.com", "exportowner3")
	projID := seedProject(t, "Export Project", "export-proj3", ownerID)

	h := NewProjectExportHandler(testDB)
	app := newTestApp("GET", "/projects/:id/export/:langCode", h.ExportLanguage, ownerID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/export/zh", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestExportLanguage_NotOwner(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "exportowner4@test.com", "exportowner4")
	outsiderID := seedUser(t, "exportoutsider@test.com", "exportoutsider")
	projID := seedProject(t, "Export Project", "export-proj4", ownerID)
	seedLanguage(t, projID, "en", "English", true)

	h := NewProjectExportHandler(testDB)
	app := newTestApp("GET", "/projects/:id/export/:langCode", h.ExportLanguage, outsiderID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/export/en", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner, got %d", resp.StatusCode)
	}
}

// contains is a simple string substring check for test assertions.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
