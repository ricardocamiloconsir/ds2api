package monitor

import (
	"context"
	"testing"
	"time"

	"ds2api/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNotifier_Subscribe(t *testing.T) {
	notifier := NewNotifier()
	ctx, cancel := context.WithCancel(context.Background())

	sub := notifier.Subscribe(ctx)
	assert.NotNil(t, sub)

	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestNotifier_NotifyExpiring(t *testing.T) {
	notifier := NewNotifier()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := notifier.Subscribe(ctx)
	go func() {
		keys := []config.APIKeyMetadata{
			{Key: "sk-expiring-1", ExpiresAt: time.Now().Add(3 * 24 * time.Hour)},
		}
		notifier.notifyExpiring(keys)
	}()

	select {
	case notification := <-sub:
		assert.Equal(t, config.NotificationTypeWarning, notification.Type)
		assert.Contains(t, notification.APIKey, "****")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive notification")
	}
}

func TestNotifier_NotifyExpired(t *testing.T) {
	notifier := NewNotifier()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := notifier.Subscribe(ctx)
	go func() {
		keys := []config.APIKeyMetadata{
			{Key: "sk-expired-1", ExpiresAt: time.Now().Add(-time.Hour)},
		}
		notifier.notifyExpired(keys)
	}()

	select {
	case notification := <-sub:
		assert.Equal(t, config.NotificationTypeExpired, notification.Type)
		assert.Contains(t, notification.APIKey, "****")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive notification")
	}
}

func TestNotifier_GetHistory(t *testing.T) {
	notifier := NewNotifier()

	keys := []config.APIKeyMetadata{
		{Key: "sk-expiring-1", ExpiresAt: time.Now().Add(3 * 24 * time.Hour)},
	}
	notifier.notifyExpiring(keys)

	history := notifier.GetHistory()
	assert.Equal(t, 1, len(history))
	assert.Equal(t, config.NotificationTypeWarning, history[0].Type)
}

func TestMonitor_GetStatus(t *testing.T) {
	store := config.NewStore(nil, "")
	store.Update(func(c *config.Config) error {
		c.Keys = []string{}
		c.APIKeys = []config.APIKeyMetadata{}
		return nil
	})

	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	status := monitor.GetStatus()
	assert.Equal(t, false, status["running"])
	assert.Equal(t, config.DefaultWarningDays, status["warning_days"])
	assert.Equal(t, config.DefaultCheckInterval.String(), status["check_interval"])
}

func TestMonitor_CheckExpirations(t *testing.T) {
	store := config.NewStore(nil, "")
	store.Update(func(c *config.Config) error {
		c.Keys = []string{}
		c.APIKeys = []config.APIKeyMetadata{}
		return nil
	})

	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	now := time.Now()

	store.Update(func(c *config.Config) error {
		c.APIKeys = []config.APIKeyMetadata{
			{Key: "sk-valid", CreatedAt: now, ExpiresAt: now.Add(config.APIKeyTTL)},
			{Key: "sk-expiring-5", CreatedAt: now, ExpiresAt: now.Add(5 * 24 * time.Hour)},
			{Key: "sk-expired", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
		}
		return nil
	})

	monitor.CheckNow()

	history := notifier.GetHistory()
	assert.GreaterOrEqual(t, len(history), 2)
}

func TestMonitor_SetWarningDays(t *testing.T) {
	store := config.NewStore(nil, "")
	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	monitor.SetWarningDays(14)

	status := monitor.GetStatus()
	assert.Equal(t, 14, status["warning_days"])
}

func TestMonitor_SetCheckInterval(t *testing.T) {
	store := config.NewStore(nil, "")
	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	interval := 12 * time.Hour
	monitor.SetCheckInterval(interval)

	status := monitor.GetStatus()
	assert.Equal(t, interval.String(), status["check_interval"])
}
