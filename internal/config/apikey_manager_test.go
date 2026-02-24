package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, err)

	metadata, found := manager.GetAPIKeyMetadata(testKey)
	require.True(t, found)
	assert.Equal(t, testKey, metadata.Key)
	assert.NotEmpty(t, metadata.ID)
	assert.NotZero(t, metadata.CreatedAt)
	assert.NotZero(t, metadata.ExpiresAt)
	assert.Equal(t, metadata.ExpiresAt.Sub(metadata.CreatedAt), APIKeyTTL)
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
	require.NoError(t, err1)

	err2 := manager.AddAPIKey(testKey)
	require.NoError(t, err2)

	cfg := store.Snapshot()
	assert.Equal(t, 1, len(cfg.APIKeys))
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
	require.NoError(t, err)

	cfg := store.Snapshot()
	assert.Equal(t, 0, len(cfg.APIKeys))
}

func TestAPIKeyManager_RemoveNonExistentKey(t *testing.T) {
	store := NewStore(nil, "")
	manager := NewAPIKeyManager(store)

	err := manager.RemoveAPIKey("sk-non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrAPIKeyNotFound, err)
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
	expiredKey := "sk-expired-key"
	manager.AddAPIKey(validKey)

	store.Update(func(c *Config) error {
		for i, metadata := range c.APIKeys {
			if metadata.Key == expiredKey {
				c.APIKeys[i].ExpiresAt = time.Now().Add(-time.Hour)
			}
		}
		return nil
	})

	assert.True(t, manager.IsAPIKeyValid(validKey))
	assert.False(t, manager.IsAPIKeyValid(expiredKey))
	assert.False(t, manager.IsAPIKeyValid("sk-non-existent"))
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
	assert.Equal(t, 2, len(expiring))
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
	assert.Equal(t, 2, len(expired))
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
	require.NoError(t, err)
	assert.Equal(t, 2, removed)

	cfg := store.Snapshot()
	assert.Equal(t, 1, len(cfg.APIKeys))
	assert.Equal(t, "sk-valid", cfg.APIKeys[0].Key)
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
	assert.Contains(t, validKeys, "sk-legacy-key")
	assert.Contains(t, validKeys, "sk-valid-1")
	assert.NotContains(t, validKeys, "sk-expired")
}

func TestGenerateAPIKeyID(t *testing.T) {
	key1 := "sk-test-key-12345"
	key2 := "sk-test-key-12345"
	key3 := "sk-different-key"

	id1 := generateAPIKeyID(key1)
	id2 := generateAPIKeyID(key2)
	id3 := generateAPIKeyID(key3)

	assert.Equal(t, id1, id2)
	assert.NotEqual(t, id1, id3)
	assert.Contains(t, id1, "apikey:")
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "long key",
			key:      "sk-abcdefghijklmnopqrstuvwxyz1234567890",
			expected: "sk-abcdefgh****7890",
		},
		{
			name:     "short key",
			key:      "sk-1234",
			expected: "****",
		},
		{
			name:     "exactly 16 chars",
			key:      "sk-12345678901234",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
