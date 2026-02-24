package admin

import (
	"context"
	"encoding/json"
	"net/http"
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

func (h *Handler) getAPIKeysMetadata(w http.ResponseWriter, r *http.Request) {
	if h.APIKeyManager == nil {
		http.Error(w, "API Key Manager not available", http.StatusServiceUnavailable)
		return
	}

	keys := h.APIKeyManager.GetAllAPIKeysMetadata()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) getExpiringKeys(w http.ResponseWriter, r *http.Request) {
	if h.APIKeyManager == nil {
		http.Error(w, "API Key Manager not available", http.StatusServiceUnavailable)
		return
	}

	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	keys := h.APIKeyManager.GetExpiringKeys(days)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) getExpiredKeys(w http.ResponseWriter, r *http.Request) {
	if h.APIKeyManager == nil {
		http.Error(w, "API Key Manager not available", http.StatusServiceUnavailable)
		return
	}

	keys := h.APIKeyManager.GetExpiredKeys()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) getNotifications(w http.ResponseWriter, r *http.Request) {
	if h.Notifier == nil {
		http.Error(w, "Notifier not available", http.StatusServiceUnavailable)
		return
	}

	notifications := h.Notifier.GetHistory()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) streamNotifications(w http.ResponseWriter, r *http.Request) {
	if h.Notifier == nil {
		http.Error(w, "Notifier not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Hour*2)
	defer cancel()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sub := h.Notifier.Subscribe(ctx)
	defer func() {
		<-ctx.Done()
		config.Logger.Debug("[sse] stream notification handler closed")
	}()

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

			if _, err := w.Write([]byte("data: ")); err != nil {
				config.Logger.Error("[sse] failed to write to stream", "error", err)
				return
			}

			if _, err := w.Write(data); err != nil {
				config.Logger.Error("[sse] failed to write to stream", "error", err)
				return
			}

			if _, err := w.Write([]byte("\n\n")); err != nil {
				config.Logger.Error("[sse] failed to write to stream", "error", err)
				return
			}

			flusher.Flush()
		}
}
}

func (h *Handler) getMonitorStatus(w http.ResponseWriter, r *http.Request) {
	if h.Monitor == nil {
		http.Error(w, "Monitor not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := h.Monitor.GetStatus()
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) updateMonitorSettings(w http.ResponseWriter, r *http.Request) {
	if h.Monitor == nil {
		http.Error(w, "Monitor not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		CheckInterval string `json:"check_interval"`
		WarningDays  int    `json:"warning_days"`
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
		h.Monitor.SetCheckInterval(interval)
	}

	if req.WarningDays > 0 {
		h.Monitor.SetWarningDays(req.WarningDays)
	}

	h.Monitor.CheckNow()

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) checkMonitorNow(w http.ResponseWriter, r *http.Request) {
	if h.Monitor == nil {
		http.Error(w, "Monitor not available", http.StatusServiceUnavailable)
		return
	}

	h.Monitor.CheckNow()

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "checked"}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
