package admin

import (
	"net/http"
	"slices"
	"strings"
)

func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
	snap := h.Store.Snapshot()
	keys := slices.Clone(snap.Keys)
	if h.APIKeyManager != nil {
		keys = h.APIKeyManager.GetValidKeys()
	} else if len(snap.APIKeys) > 0 {
		seen := make(map[string]struct{}, len(snap.Keys)+len(snap.APIKeys))
		keys = make([]string, 0, len(snap.Keys)+len(snap.APIKeys))

		for _, key := range snap.Keys {
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}

		for _, metadata := range snap.APIKeys {
			if _, exists := seen[metadata.Key]; exists {
				continue
			}
			seen[metadata.Key] = struct{}{}
			keys = append(keys, metadata.Key)
		}
	}
	safe := map[string]any{
		"keys":     keys,
		"accounts": []map[string]any{},
		"claude_mapping": func() map[string]string {
			if len(snap.ClaudeMapping) > 0 {
				return snap.ClaudeMapping
			}
			return snap.ClaudeModelMap
		}(),
	}
	accounts := make([]map[string]any, 0, len(snap.Accounts))
	for _, acc := range snap.Accounts {
		token := strings.TrimSpace(acc.Token)
		preview := ""
		if token != "" {
			if len(token) > 20 {
				preview = token[:20] + "..."
			} else {
				preview = token
			}
		}
		accounts = append(accounts, map[string]any{
			"identifier":    acc.Identifier(),
			"email":         acc.Email,
			"mobile":        acc.Mobile,
			"has_password":  strings.TrimSpace(acc.Password) != "",
			"has_token":     token != "",
			"token_preview": preview,
		})
	}
	safe["accounts"] = accounts
	writeJSON(w, http.StatusOK, safe)
}

func (h *Handler) exportConfig(w http.ResponseWriter, _ *http.Request) {
	h.configExport(w, nil)
}

func (h *Handler) configExport(w http.ResponseWriter, _ *http.Request) {
	snap := h.Store.Snapshot()
	jsonStr, b64, err := h.Store.ExportJSONAndBase64()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"config":  snap,
		"json":    jsonStr,
		"base64":  b64,
	})
}
