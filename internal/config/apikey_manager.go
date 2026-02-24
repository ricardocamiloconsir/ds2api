package config

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"time"
)

type APIKeyManager struct {
	store *Store
}

func NewAPIKeyManager(store *Store) *APIKeyManager {
	return &APIKeyManager{store: store}
}

func (m *APIKeyManager) AddAPIKey(key string) error {
	if key == "" {
		return ErrInvalidAPIKey
	}

	now := time.Now()
	metadata := APIKeyMetadata{
		ID:        generateAPIKeyID(key),
		Key:       key,
		CreatedAt: now,
		ExpiresAt: now.Add(APIKeyTTL),
	}

	return m.store.Update(func(c *Config) error {
		for i, existing := range c.APIKeys {
			if existing.Key == key {
				c.APIKeys[i] = metadata
				return nil
			}
		}
		c.APIKeys = append(c.APIKeys, metadata)
		return nil
	})
}

func (m *APIKeyManager) RemoveAPIKey(key string) error {
	return m.store.Update(func(c *Config) error {
		for i, metadata := range c.APIKeys {
			if metadata.Key == key {
				c.APIKeys = append(c.APIKeys[:i], c.APIKeys[i+1:]...)
				return nil
			}
		}
		return ErrAPIKeyNotFound
	})
}

type KeyFilterFunc func(APIKeyMetadata) bool

func (m *APIKeyManager) filterKeys(filter KeyFilterFunc) []APIKeyMetadata {
	cfg := m.store.Snapshot()
	result := make([]APIKeyMetadata, 0)
	for _, metadata := range cfg.APIKeys {
		if filter(metadata) {
			result = append(result, metadata)
		}
	}
	return result
}

func (m *APIKeyManager) IsAPIKeyValid(key string) bool {
	now := time.Now()
	cfg := m.store.Snapshot()
	for _, metadata := range cfg.APIKeys {
		if metadata.Key == key {
			return now.Before(metadata.ExpiresAt)
		}
	}

	for _, k := range cfg.Keys {
		if k == key {
			return true
		}
	}

	return false
}

func (m *APIKeyManager) GetAPIKeyMetadata(key string) (APIKeyMetadata, bool) {
	cfg := m.store.Snapshot()
	for _, metadata := range cfg.APIKeys {
		if metadata.Key == key {
			return metadata, true
		}
	}

	return APIKeyMetadata{}, false
}

func (m *APIKeyManager) GetExpiringKeys(daysBefore int) []APIKeyMetadata {
	now := time.Now()
	threshold := now.Add(time.Duration(daysBefore) * 24 * time.Hour)
	return m.filterKeys(func(k APIKeyMetadata) bool {
		return k.ExpiresAt.After(now) && k.ExpiresAt.Before(threshold)
	})
}

func (m *APIKeyManager) GetExpiredKeys() []APIKeyMetadata {
	now := time.Now()
	return m.filterKeys(func(k APIKeyMetadata) bool {
		return k.ExpiresAt.Before(now)
	})
}

func (m *APIKeyManager) CleanExpiredKeys() (int, error) {
	var removed int
	err := m.store.Update(func(c *Config) error {
		var valid []APIKeyMetadata
		now := time.Now()
		for _, metadata := range c.APIKeys {
			if metadata.ExpiresAt.After(now) {
				valid = append(valid, metadata)
			} else {
				removed++
			}
		}
		c.APIKeys = valid
		return nil
	})

	if err != nil {
		return 0, err
	}

	return removed, nil
}

func (m *APIKeyManager) GetAllAPIKeysMetadata() []APIKeyMetadata {
	cfg := m.store.Snapshot()
	return slices.Clone(cfg.APIKeys)
}

func (m *APIKeyManager) GetValidKeys() []string {
	cfg := m.store.Snapshot()
	now := time.Now()
	validKeys := make([]string, 0, len(cfg.APIKeys)+len(cfg.Keys))
	seen := make(map[string]struct{}, len(cfg.APIKeys)+len(cfg.Keys))

	for _, metadata := range cfg.APIKeys {
		if metadata.ExpiresAt.After(now) {
			if _, exists := seen[metadata.Key]; !exists {
				seen[metadata.Key] = struct{}{}
				validKeys = append(validKeys, metadata.Key)
			}
		}
	}

	for _, k := range cfg.Keys {
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			validKeys = append(validKeys, k)
		}
	}

	return validKeys
}

func (m *APIKeyManager) GetValidAPIKeysMetadata() []APIKeyMetadata {
	cfg := m.store.Snapshot()
	now := time.Now()
	validMetadata := make([]APIKeyMetadata, 0, len(cfg.APIKeys))

	for _, metadata := range cfg.APIKeys {
		if metadata.ExpiresAt.After(now) {
			validMetadata = append(validMetadata, metadata)
		}
	}

	return validMetadata
}

func generateAPIKeyID(key string) string {
	sum := sha256.Sum256([]byte(key))
	return "apikey:" + hex.EncodeToString(sum[:8])
}

var (
	ErrInvalidAPIKey  = newConfigError("invalid_api_key", "API key cannot be empty")
	ErrAPIKeyNotFound = newConfigError("api_key_not_found", "API key not found")
	ErrAPIKeyExpired  = newConfigError("api_key_expired", "API key has expired")
	ErrAPIKeyExpiring = newConfigError("api_key_expiring", "API key is expiring soon")
)

type ConfigError struct {
	Code    string
	Message string
}

func newConfigError(code, message string) *ConfigError {
	return &ConfigError{Code: code, Message: message}
}

func (e *ConfigError) Error() string {
	return e.Message
}

func (e *ConfigError) CodeStr() string {
	return e.Code
}

func maskAPIKey(key string) string {
	if len(key) <= 17 {
		return "****"
	}
	return key[:11] + "****" + key[len(key)-4:]
}
