package config

import (
	"slices"
	"testing"
	"time"
)

func TestAPIKeyManager_AddAPIKey(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	testKey := "sk-test-key-12345"
	err := manager.AddAPIKey(testKey)
	if err != nil {
		t.Fatalf("AddAPIKey returned error: %v", err)
	}

	metadata, found := manager.GetAPIKeyMetadata(testKey)
	if !found {
		t.Fatal("expected key metadata to be found")
	}
	if metadata.Key != testKey {
		t.Fatalf("expected key %q, got %q", testKey, metadata.Key)
	}
	if metadata.ID == "" {
		t.Fatal("expected metadata ID to be set")
	}
	if metadata.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
	if !metadata.ExpiresAt.IsZero() {
		t.Fatal("expected ExpiresAt to be empty")
	}
}

func TestAPIKeyManager_AddDuplicateKey(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	testKey := "sk-test-key-12345"

	err1 := manager.AddAPIKey(testKey)
	if err1 != nil {
		t.Fatalf("first AddAPIKey returned error: %v", err1)
	}

	err2 := manager.AddAPIKey(testKey)
	if err2 != nil {
		t.Fatalf("second AddAPIKey returned error: %v", err2)
	}

	cfg := store.Snapshot()
	if len(cfg.APIKeys) != 1 {
		t.Fatalf("expected 1 API key, got %d", len(cfg.APIKeys))
	}
}

func TestAPIKeyManager_RemoveAPIKey(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	testKey := "sk-test-key-12345"
	manager.AddAPIKey(testKey)

	err := manager.RemoveAPIKey(testKey)
	if err != nil {
		t.Fatalf("RemoveAPIKey returned error: %v", err)
	}

	cfg := store.Snapshot()
	if len(cfg.APIKeys) != 0 {
		t.Fatalf("expected 0 API keys, got %d", len(cfg.APIKeys))
	}
}

func TestAPIKeyManager_RemoveNonExistentKey(t *testing.T) {
	store := NewStore(nil, "")
	manager := NewAPIKeyManager(store)

	err := manager.RemoveAPIKey("sk-non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if err != ErrAPIKeyNotFound {
		t.Fatalf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestAPIKeyManager_IsAPIKeyValid(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	validKey := "sk-valid-key"
	manager.AddAPIKey(validKey)

	if !manager.IsAPIKeyValid(validKey) {
		t.Fatal("expected valid key to be valid")
	}
	if manager.IsAPIKeyValid("sk-non-existent") {
		t.Fatal("expected non-existent key to be invalid")
	}
}

func TestAPIKeyManager_GetExpiringKeys(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	now := time.Now()

	store.Update(func(c *Config) error {
		c.APIKeys = []APIKeyMetadata{
			{Key: "sk-valid", CreatedAt: now, ExpiresAt: now.Add(APIKeyTTL)},
			{Key: "sk-expiring-5", CreatedAt: now, ExpiresAt: now.Add(5 * 24 * time.Hour)},
			{Key: "sk-expiring-3", CreatedAt: now, ExpiresAt: now.Add(3 * 24 * time.Hour)},
			{Key: "sk-expired", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
		}
		return nil
	})

	expiring := manager.GetExpiringKeys(7)
	if len(expiring) != 0 {
		t.Fatalf("expected 0 expiring keys, got %d", len(expiring))
	}
}

func TestAPIKeyManager_GetExpiredKeys(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	now := time.Now()

	store.Update(func(c *Config) error {
		c.APIKeys = []APIKeyMetadata{
			{Key: "sk-valid", CreatedAt: now, ExpiresAt: now.Add(APIKeyTTL)},
			{Key: "sk-expired-1", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
			{Key: "sk-expired-2", CreatedAt: now, ExpiresAt: now.Add(-2 * time.Hour)},
		}
		return nil
	})

	expired := manager.GetExpiredKeys()
	if len(expired) != 0 {
		t.Fatalf("expected 0 expired keys, got %d", len(expired))
	}
}

func TestAPIKeyManager_CleanExpiredKeys(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	now := time.Now()

	store.Update(func(c *Config) error {
		c.APIKeys = []APIKeyMetadata{
			{Key: "sk-valid", CreatedAt: now, ExpiresAt: now.Add(APIKeyTTL)},
			{Key: "sk-expired-1", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
			{Key: "sk-expired-2", CreatedAt: now, ExpiresAt: now.Add(-2 * time.Hour)},
		}
		return nil
	})

	removed, err := manager.CleanExpiredKeys()
	if err != nil {
		t.Fatalf("CleanExpiredKeys returned error: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed keys, got %d", removed)
	}
}

func TestAPIKeyManager_GetValidKeys(t *testing.T) {
	store := NewStore(nil, "")
	store.Update(func(c *Config) error {
		c.Keys = []string{"sk-legacy-key"}
		c.APIKeys = []APIKeyMetadata{}
		return nil
	})

	manager := NewAPIKeyManager(store)

	now := time.Now()

	store.Update(func(c *Config) error {
		c.APIKeys = []APIKeyMetadata{
			{Key: "sk-valid-1", CreatedAt: now, ExpiresAt: now.Add(APIKeyTTL)},
			{Key: "sk-expired", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
		}
		return nil
	})

	validKeys := manager.GetValidKeys()
	if !slices.Contains(validKeys, "sk-legacy-key") {
		t.Fatal("expected legacy key in valid keys")
	}
	if !slices.Contains(validKeys, "sk-valid-1") {
		t.Fatal("expected valid metadata key in valid keys")
	}
	if !slices.Contains(validKeys, "sk-expired") {
		t.Fatal("expected metadata key in valid keys")
	}
}
