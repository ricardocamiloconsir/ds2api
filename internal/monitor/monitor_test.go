package monitor

import (
	"context"
	"strings"
	"testing"
	"time"

	"ds2api/internal/config"
)

func TestNotifier_Subscribe(t *testing.T) {
	notifier := NewNotifier()
	ctx, cancel := context.WithCancel(context.Background())

	sub := notifier.Subscribe(ctx)
	if sub == nil {
		t.Fatal("expected subscriber channel to be non-nil")
	}

	cancel()

	deadline := time.After(200 * time.Millisecond)
	for {
		select {
		case _, ok := <-sub:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("subscriber channel was not closed after cancel")
		}
	}
}

func TestNotifier_NotifyExpiring(t *testing.T) {
	notifier := NewNotifier()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sub := notifier.Subscribe(ctx)
	go func() {
		keys := []config.APIKeyMetadata{{Key: "sk-expiring-1", ExpiresAt: time.Now().Add(3 * 24 * time.Hour)}}
		notifier.notifyExpiring(keys)
	}()

	select {
	case notification := <-sub:
		if notification.Type != config.NotificationTypeWarning {
			t.Fatalf("expected type %q, got %q", config.NotificationTypeWarning, notification.Type)
		}
		if !strings.Contains(notification.APIKey, "****") {
			t.Fatalf("expected obfuscated API key, got %q", notification.APIKey)
		}
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
		keys := []config.APIKeyMetadata{{Key: "sk-expired-1", ExpiresAt: time.Now().Add(-time.Hour)}}
		notifier.notifyExpired(keys)
	}()

	select {
	case notification := <-sub:
		if notification.Type != config.NotificationTypeExpired {
			t.Fatalf("expected type %q, got %q", config.NotificationTypeExpired, notification.Type)
		}
		if !strings.Contains(notification.APIKey, "****") {
			t.Fatalf("expected obfuscated API key, got %q", notification.APIKey)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive notification")
	}
}

func TestNotifier_GetHistory(t *testing.T) {
	notifier := NewNotifier()

	keys := []config.APIKeyMetadata{{Key: "sk-expiring-1", ExpiresAt: time.Now().Add(3 * 24 * time.Hour)}}
	notifier.notifyExpiring(keys)

	history := notifier.GetHistory()
	if len(history) != 1 {
		t.Fatalf("expected history len 1, got %d", len(history))
	}
	if history[0].Type != config.NotificationTypeWarning {
		t.Fatalf("expected history notification type %q, got %q", config.NotificationTypeWarning, history[0].Type)
	}
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
	if status["running"] != false {
		t.Fatalf("expected running false, got %#v", status["running"])
	}
	if status["warning_days"] != config.DefaultWarningDays {
		t.Fatalf("expected warning_days %d, got %#v", config.DefaultWarningDays, status["warning_days"])
	}
	if status["check_interval"] != config.DefaultCheckInterval.String() {
		t.Fatalf("expected check_interval %q, got %#v", config.DefaultCheckInterval.String(), status["check_interval"])
	}
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
	if len(history) != 0 {
		t.Fatalf("expected 0 notifications, got %d", len(history))
	}
}

func TestMonitor_SetWarningDays(t *testing.T) {
	store := config.NewStore(nil, "")
	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	monitor.SetWarningDays(14)

	status := monitor.GetStatus()
	if status["warning_days"] != 14 {
		t.Fatalf("expected warning_days 14, got %#v", status["warning_days"])
	}
}

func TestMonitor_SetCheckInterval(t *testing.T) {
	store := config.NewStore(nil, "")
	apiKeyManager := config.NewAPIKeyManager(store)
	notifier := NewNotifier()
	monitor := NewMonitor(store, apiKeyManager, notifier)

	interval := 12 * time.Hour
	monitor.SetCheckInterval(interval)

	status := monitor.GetStatus()
	if status["check_interval"] != interval.String() {
		t.Fatalf("expected check_interval %q, got %#v", interval.String(), status["check_interval"])
	}
}
