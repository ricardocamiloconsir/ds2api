package monitor

import (
	"context"
	"sync"
	"time"

	"ds2api/internal/config"
)

type Monitor struct {
	store         *config.Store
	apiKeyManager *config.APIKeyManager
	notifier      *Notifier
	checkInterval time.Duration
	warningDays   int
	cancel        context.CancelFunc
	running       bool
	mu            sync.Mutex
}

func NewMonitor(store *config.Store, apiKeyManager *config.APIKeyManager, notifier *Notifier) *Monitor {
	return &Monitor{
		store:         store,
		apiKeyManager: apiKeyManager,
		notifier:      notifier,
		checkInterval: config.DefaultCheckInterval,
		warningDays:   config.DefaultWarningDays,
	}
}

func (m *Monitor) SetCheckInterval(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkInterval = interval
}

func (m *Monitor) SetWarningDays(days int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warningDays = days
}

func (m *Monitor) Start(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	config.Logger.Info("[monitor] starting API key expiration monitor", "interval", m.checkInterval, "warningDays", m.warningDays)

	m.checkExpirations()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			config.Logger.Info("[monitor] stopping API key expiration monitor")
			return
		case <-ticker.C:
			m.checkExpirations()
		}
	}
}

func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
	}
	m.running = false
}

func (m *Monitor) CheckNow() {
	m.checkExpirations()
}

func (m *Monitor) checkExpirations() {
	config.Logger.Debug("[monitor] checking API key expirations")

	expiring := m.apiKeyManager.GetExpiringKeys(m.warningDays)
	if len(expiring) > 0 {
		config.Logger.Info("[monitor] found expiring API keys", "count", len(expiring), "days", m.warningDays)
		m.notifier.notifyExpiring(expiring)
	}

	expired := m.apiKeyManager.GetExpiredKeys()
	if len(expired) > 0 {
		config.Logger.Info("[monitor] found expired API keys", "count", len(expired))
		m.notifier.notifyExpired(expired)
	}

	if len(expiring) == 0 && len(expired) == 0 {
		config.Logger.Debug("[monitor] no expiring or expired API keys found")
	}
}

func (m *Monitor) GetStatus() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiring := m.apiKeyManager.GetExpiringKeys(m.warningDays)
	expired := m.apiKeyManager.GetExpiredKeys()
	allKeys := m.apiKeyManager.GetAllAPIKeysMetadata()

	return map[string]any{
		"running":        m.running,
		"check_interval":  m.checkInterval.String(),
		"warning_days":   m.warningDays,
		"total_keys":     len(allKeys),
		"expiring_keys":  len(expiring),
		"expired_keys":   len(expired),
		"valid_keys":     len(allKeys) - len(expired),
	}
}
