package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ds2api/internal/config"
)

func TestAddKeyWithManagerPersistsAndListedInConfig(t *testing.T) {
	h := newAdminTestHandler(t, `{}`)
	h.APIKeyManager = config.NewAPIKeyManager(h.Store.(*config.Store))

	addReq := httptest.NewRequest(http.MethodPost, "/admin/keys", strings.NewReader(`{"key":"sk-test-123"}`))
	addRec := httptest.NewRecorder()
	h.addKey(addRec, addReq)
	if addRec.Code != http.StatusOK {
		t.Fatalf("unexpected add status: %d body=%s", addRec.Code, addRec.Body.String())
	}

	cfgReq := httptest.NewRequest(http.MethodGet, "/admin/config", nil)
	cfgRec := httptest.NewRecorder()
	h.getConfig(cfgRec, cfgReq)
	if cfgRec.Code != http.StatusOK {
		t.Fatalf("unexpected config status: %d body=%s", cfgRec.Code, cfgRec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(cfgRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	keys, _ := payload["keys"].([]any)
	if len(keys) != 1 || keys[0] != "sk-test-123" {
		t.Fatalf("unexpected keys list: %#v", keys)
	}
}

func TestAddKeyRejectsInvalidJSON(t *testing.T) {
	h := newAdminTestHandler(t, `{}`)

	req := httptest.NewRequest(http.MethodPost, "/admin/keys", strings.NewReader(`{"key":`))
	rec := httptest.NewRecorder()
	h.addKey(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}
}
