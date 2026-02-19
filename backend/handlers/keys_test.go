package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"translate-management/models"
)

// ---- Keys Handler Tests ----

func TestListKeys(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner@test.com", "keyowner")
	projID := seedProject(t, "Key Project", "key-proj", ownerID)
	seedKey(t, projID, "home.title", "Home title")
	seedKey(t, projID, "home.subtitle", "Home subtitle")

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("GET", "/projects/:id/keys", h.List, ownerID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/keys", projID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var keys []models.TranslationKey
	json.NewDecoder(resp.Body).Decode(&keys)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestListKeys_NotMember(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner2@test.com", "keyowner2")
	outsiderID := seedUser(t, "outsider@test.com", "outsider")
	projID := seedProject(t, "Key Project", "key-proj2", ownerID)

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("GET", "/projects/:id/keys", h.List, outsiderID)

	resp := doRequest(t, app, "GET", fmt.Sprintf("/projects/%s/keys", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateKey(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner3@test.com", "keyowner3")
	projID := seedProject(t, "Key Project", "key-proj3", ownerID)

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("POST", "/projects/:id/keys", h.Create, ownerID)

	body := map[string]string{"key": "nav.home", "description": "Navigation home"}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/keys", projID), body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var k models.TranslationKey
	json.NewDecoder(resp.Body).Decode(&k)
	if k.Key != "nav.home" {
		t.Errorf("expected key 'nav.home', got '%s'", k.Key)
	}
}

func TestCreateKey_DuplicateKey(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner4@test.com", "keyowner4")
	projID := seedProject(t, "Key Project", "key-proj4", ownerID)
	seedKey(t, projID, "nav.home", "existing")

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("POST", "/projects/:id/keys", h.Create, ownerID)

	body := map[string]string{"key": "nav.home"}
	resp := doRequest(t, app, "POST", fmt.Sprintf("/projects/%s/keys", projID), body)
	if resp.StatusCode == http.StatusCreated {
		t.Fatal("expected error for duplicate key, got 201")
	}
}

func TestDeleteKey(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner5@test.com", "keyowner5")
	projID := seedProject(t, "Key Project", "key-proj5", ownerID)
	keyID := seedKey(t, projID, "nav.about", "")

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("DELETE", "/projects/:id/keys/:keyId", h.Delete, ownerID)

	resp := doRequest(t, app, "DELETE", fmt.Sprintf("/projects/%s/keys/%s", projID, keyID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteKey_NotFound(t *testing.T) {
	cleanDB(t)
	ownerID := seedUser(t, "keyowner6@test.com", "keyowner6")
	projID := seedProject(t, "Key Project", "key-proj6", ownerID)

	h := NewKeyHandler(testDB, nil)
	app := newTestApp("DELETE", "/projects/:id/keys/:keyId", h.Delete, ownerID)

	resp := doRequest(t, app, "DELETE",
		fmt.Sprintf("/projects/%s/keys/00000000-0000-0000-0000-000000000000", projID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
