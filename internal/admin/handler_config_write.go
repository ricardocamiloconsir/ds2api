package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ds2api/internal/config"
)

func (h *Handler) updateConfig(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid json"})
		return
	}
	old := h.Store.Snapshot()
	err := h.Store.Update(func(c *config.Config) error {
		if keys, ok := toStringSlice(req["keys"]); ok {
			c.Keys = keys
		}
		if accountsRaw, ok := req["accounts"].([]any); ok {
			existing := map[string]config.Account{}
			for _, a := range old.Accounts {
				existing[a.Identifier()] = a
			}
			accounts := make([]config.Account, 0, len(accountsRaw))
			for _, item := range accountsRaw {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				acc := toAccount(m)
				id := acc.Identifier()
				if prev, ok := existing[id]; ok {
					if strings.TrimSpace(acc.Password) == "" {
						acc.Password = prev.Password
					}
					if strings.TrimSpace(acc.Token) == "" {
						acc.Token = prev.Token
					}
				}
				accounts = append(accounts, acc)
			}
			c.Accounts = accounts
		}
		if m, ok := req["claude_mapping"].(map[string]any); ok {
			newMap := map[string]string{}
			for k, v := range m {
				newMap[k] = fmt.Sprintf("%v", v)
			}
			c.ClaudeMapping = newMap
		}
		return nil
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}
	h.Pool.Reset()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "配置已更新"})
}

func (h *Handler) addKey(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		config.Logger.Error("[admin][keys] failed to decode add key request", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "invalid json"})
		return
	}
	key, _ := req["key"].(string)
	key = strings.TrimSpace(key)
	if key == "" {
		config.Logger.Warn("[admin][keys] rejected empty key")
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "Key 不能为空"})
		return
	}

	masked := safeTruncate(key, 8)
	config.Logger.Info("[admin][keys] add key requested", "key", masked, "has_manager", h.APIKeyManager != nil)

	if h.APIKeyManager != nil {
		if err := h.APIKeyManager.AddAPIKey(key); err != nil {
			config.Logger.Error("[admin][keys] failed to persist key via manager", "key", masked, "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
			return
		}
		totalKeys := len(h.APIKeyManager.GetValidKeys())
		config.Logger.Info("[admin][keys] key persisted via manager", "key", masked, "total_keys", totalKeys)
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "total_keys": totalKeys})
		return
	}

	err := h.Store.Update(func(c *config.Config) error {
		for _, k := range c.Keys {
			if k == key {
				return fmt.Errorf("Key 已存在")
			}
		}
		c.Keys = append(c.Keys, key)
		return nil
	})
	if err != nil {
		config.Logger.Error("[admin][keys] failed to persist key via store", "key", masked, "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": err.Error()})
		return
	}
	totalKeys := len(h.Store.Snapshot().Keys)
	config.Logger.Info("[admin][keys] key persisted via store", "key", masked, "total_keys", totalKeys)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "total_keys": totalKeys})
}

func (h *Handler) deleteKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	if h.APIKeyManager != nil {
		if err := h.APIKeyManager.RemoveAPIKey(key); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"detail": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "total_keys": len(h.APIKeyManager.GetValidKeys())})
		return
	}

	err := h.Store.Update(func(c *config.Config) error {
		idx := -1
		for i, k := range c.Keys {
			if k == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("Key 不存在")
		}
		c.Keys = append(c.Keys[:idx], c.Keys[idx+1:]...)
		return nil
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "total_keys": len(h.Store.Snapshot().Keys)})
}

func (h *Handler) batchImport(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"detail": "无效的 JSON 格式"})
		return
	}
	importedKeys, importedAccounts := 0, 0
	err := h.Store.Update(func(c *config.Config) error {
		if keys, ok := req["keys"].([]any); ok {
			existing := map[string]bool{}
			for _, k := range c.Keys {
				existing[k] = true
			}
			for _, k := range keys {
				key := strings.TrimSpace(fmt.Sprintf("%v", k))
				if key == "" || existing[key] {
					continue
				}
				c.Keys = append(c.Keys, key)
				existing[key] = true
				importedKeys++
			}
		}
		if accounts, ok := req["accounts"].([]any); ok {
			existing := map[string]bool{}
			for _, a := range c.Accounts {
				existing[a.Identifier()] = true
			}
			for _, item := range accounts {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}
				acc := toAccount(m)
				id := acc.Identifier()
				if id == "" || existing[id] {
					continue
				}
				c.Accounts = append(c.Accounts, acc)
				existing[id] = true
				importedAccounts++
			}
		}
		return nil
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"detail": err.Error()})
		return
	}
	h.Pool.Reset()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "imported_keys": importedKeys, "imported_accounts": importedAccounts})
}
