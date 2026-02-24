package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"ds2api/internal/config"
	"ds2api/internal/monitor"
)

type Handler struct {
	Store         ConfigStore
	Pool          PoolController
	DS            DeepSeekCaller
	APIKeyManager *config.APIKeyManager
	Monitor       *monitor.Monitor
	Notifier      *monitor.Notifier
}

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Post("/login", h.login)
	r.Get("/verify", h.verify)
	r.Group(func(pr chi.Router) {
		pr.Use(h.requireAdmin)
		pr.Get("/vercel/config", h.getVercelConfig)
		pr.Get("/config", h.getConfig)
		pr.Post("/config", h.updateConfig)
		pr.Get("/settings", h.getSettings)
		pr.Put("/settings", h.updateSettings)
		pr.Post("/settings/password", h.updateSettingsPassword)
		pr.Post("/config/import", h.configImport)
		pr.Get("/config/export", h.configExport)
		pr.Post("/keys", h.addKey)
		pr.Delete("/keys/{key}", h.deleteKey)
		pr.Get("/keys/metadata", h.getAPIKeysMetadata)
		pr.Get("/keys/expiring", h.getExpiringKeys)
		pr.Get("/keys/expired", h.getExpiredKeys)
		pr.Get("/accounts", h.listAccounts)
		pr.Post("/accounts", h.addAccount)
		pr.Put("/accounts/{identifier}", h.updateAccount)
		pr.Delete("/accounts/{identifier}", h.deleteAccount)
		pr.Get("/queue/status", h.queueStatus)
		pr.Post("/accounts/test", h.testSingleAccount)
		pr.Post("/accounts/test-all", h.testAllAccounts)
		pr.Post("/import", h.batchImport)
		pr.Post("/test", h.testAPI)
		pr.Post("/vercel/sync", h.syncVercel)
		pr.Get("/vercel/status", h.vercelStatus)
		pr.Get("/export", h.exportConfig)
		pr.Get("/dev/captures", h.getDevCaptures)
		pr.Delete("/dev/captures", h.clearDevCaptures)
		pr.Get("/notifications", h.getNotifications)
		pr.Get("/notifications/stream", h.streamNotifications)
		pr.Get("/monitor/status", h.getMonitorStatus)
		pr.Put("/monitor/settings", h.updateMonitorSettings)
		pr.Post("/monitor/check", h.checkMonitorNow)
	})
}

func (h *Handler) writeJSONResponse(w http.ResponseWriter, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		config.Logger.Error("[admin] failed to marshal JSON response", "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		config.Logger.Error("[admin] failed to write JSON response", "error", err)
	}
}

func requireService[T any](service T, serviceName string, w http.ResponseWriter) bool {
	v := reflect.ValueOf(service)
	if !v.IsValid() || ((v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Func || v.Kind() == reflect.Chan) && v.IsNil()) {
		http.Error(w, serviceName+" not available", http.StatusServiceUnavailable)
		return false
	}
	return true
}

func (h *Handler) writeSSEData(w http.ResponseWriter, flusher http.Flusher, data []byte) bool {
	if _, err := w.Write([]byte("data: ")); err != nil {
		config.Logger.Error("[sse] failed to write prefix", "error", err)
		return false
	}
	if _, err := w.Write(data); err != nil {
		config.Logger.Error("[sse] failed to write data", "error", err)
		return false
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		config.Logger.Error("[sse] failed to write newline", "error", err)
		return false
	}
	flusher.Flush()
	return true
}

func (h *Handler) getAPIKeysMetadata(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.APIKeyManager, "API Key Manager", w) {
		return
	}
	h.writeJSONResponse(w, h.APIKeyManager.GetAllAPIKeysMetadata())
}

func (h *Handler) getExpiringKeys(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.APIKeyManager, "API Key Manager", w) {
		return
	}

	days := config.DefaultWarningDays
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}
	h.writeJSONResponse(w, h.APIKeyManager.GetExpiringKeys(days))
}

func (h *Handler) getExpiredKeys(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.APIKeyManager, "API Key Manager", w) {
		return
	}
	h.writeJSONResponse(w, h.APIKeyManager.GetExpiredKeys())
}

func (h *Handler) getNotifications(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.Notifier, "Notifier", w) {
		return
	}
	h.writeJSONResponse(w, h.Notifier.GetHistory())
}

func (h *Handler) streamNotifications(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.Notifier, "Notifier", w) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), config.SSETimeoutDefault)
	defer cancel()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", config.SSEContentType)
	w.Header().Set("Cache-Control", config.SSECacheControl)
	w.Header().Set("Connection", config.SSEConnection)

	sub := h.Notifier.Subscribe(ctx)
	defer config.Logger.Debug("[sse] stream notification handler closed")

	for {
		select {
		case <-ctx.Done():
			config.Logger.Debug("[sse] context cancelled, closing stream")
			return
		case notification, ok := <-sub:
			if !ok {
				config.Logger.Debug("[sse] notification channel closed")
				return
			}

			data, err := json.Marshal(notification)
			if err != nil {
				config.Logger.Error("[sse] failed to marshal notification", "error", err, "notification", notification)
				continue
			}

			if !h.writeSSEData(w, flusher, data) {
				config.Logger.Debug("[sse] write failed, closing stream")
				return
			}
		}
	}
}

func (h *Handler) getMonitorStatus(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.Monitor, "Monitor", w) {
		return
	}
	h.writeJSONResponse(w, h.Monitor.GetStatus())
}

func (h *Handler) updateMonitorSettings(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.Monitor, "Monitor", w) {
		return
	}

	var req struct {
		CheckInterval string `json:"check_interval"`
		WarningDays   int    `json:"warning_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CheckInterval != "" {
		interval, err := time.ParseDuration(req.CheckInterval)
		if err != nil {
			http.Error(w, "Invalid check_interval format", http.StatusBadRequest)
			return
		}
		if interval <= 0 {
			http.Error(w, "check_interval must be a positive duration", http.StatusBadRequest)
			return
		}
		h.Monitor.SetCheckInterval(interval)
	}

	if req.WarningDays > 0 {
		h.Monitor.SetWarningDays(req.WarningDays)
	}

	h.Monitor.CheckNow()

	h.writeJSONResponse(w, map[string]string{"status": "updated"})
}

func (h *Handler) checkMonitorNow(w http.ResponseWriter, r *http.Request) {
	if !requireService(h.Monitor, "Monitor", w) {
		return
	}

	h.Monitor.CheckNow()

	h.writeJSONResponse(w, map[string]string{"status": "checked"})
}
